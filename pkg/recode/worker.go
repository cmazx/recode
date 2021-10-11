package recode

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/cmazx/recode/pkg/convert"
	"github.com/cmazx/recode/pkg/queue"
	"github.com/cmazx/recode/pkg/storage"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
)

type Worker struct {
	signalCh      chan os.Signal
	queueService  *queue.Queue
	stopCh        chan interface{}
	formatStorage *MediaFormatStorage
	jobStorage    *JobStorage
	tempMediaPath string
	jobTopicName  string
	fileStorage   storage.Storage
}

func NewWorker(queueService *queue.Queue, formatStorage *MediaFormatStorage, jobStorage *JobStorage, stg storage.Storage, signalCh chan os.Signal) *Worker {
	tempMediaPath := os.Getenv("MEDIA_TEMP_PATH")
	if tempMediaPath == "" {
		panic("No temporal media path specified in MEDIA_TEMP_PATH env variable")
	}
	log.Println("Temp media path:" + tempMediaPath)
	err := os.MkdirAll(tempMediaPath, 755)
	if err != nil {
		log.Fatalf("Error create media temporary directory: %s", err)
		return nil
	}
	return &Worker{
		signalCh,
		queueService,
		make(chan interface{}, 1),
		formatStorage,
		jobStorage,
		tempMediaPath,
		os.Getenv("WORKER_JOB_TOPIC"),
		stg,
	}
}

func (w *Worker) Start() {
	e := echo.New()
	e.GET("/stop", func(c echo.Context) error {
		fmt.Println("Exit gracefully")
		w.signalCh <- syscall.SIGTERM
		return nil
	})
	wg := sync.WaitGroup{}

	wg.Add(1)
	//rest server
	go func() {
		defer wg.Done()
		e.Logger.Fatal(e.Start(":" + os.Getenv("WORKER_PORT")))
	}()

	wg.Add(1)

	//consumer
	go func() {
		consumerTicker := time.NewTicker(500 * time.Millisecond)
		defer wg.Done()
		for {
			select {
			case <-w.stopCh:
				consumerTicker.Stop()
				return
			case <-consumerTicker.C:
				err := w.queueService.Consume(func(payload []byte) error {
					log.Println("Job message consumed " + string(payload))
					job := JobData{}
					err := json.Unmarshal(payload, &job)
					if err != nil {
						return err
					}

					w.doJob(job)
					return nil
				})
				if err != nil {
					panic(err)
				}
			}
		}
	}()

	log.Printf("Working...")
	<-w.signalCh
	w.stopCh <- true

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

	log.Printf("Waiting goroutines")
	wg.Wait()
	log.Printf("Worker stopped")
}

func (w *Worker) doJob(data JobData) {
	log.Printf("Processing job %v", data.Job)
	tempPath, err := w.fetchSource(data.Job.SourceUrl)
	if err != nil {
		data.Job.Status = Failed
		w.jobStorage.update(&data.Job)
		log.Println("Error on getting source data: " + err.Error())
		return
	}

	data.Job.Status = Running
	w.jobStorage.update(&data.Job)

	data.Job.Status = Done
	var results []*JobResult
	for _, format := range data.Formats {
		log.Printf("Processing job %d format %v\n", data.Job.ID, format)
		mf := convert.MediaFormat{
			Name:     format.Name,
			Width:    format.Width,
			Height:   format.Height,
			Quality:  format.Quality,
			Encoding: format.Encoding,
		}
		resultFilePath, err := convert.Process(mf, tempPath, w.tempMediaPath)
		data.Job.Status = Done
		details := ""
		if err != nil {
			data.Job.Status = Failed
			log.Printf("Processing job %d  error %s\n", data.Job.ID, err.Error())
			details = err.Error()
		}

		storagePath := strings.Trim(data.Job.TargetPath, "/") + "/" + filepath.Base(resultFilePath)
		publicUrl, err := w.fileStorage.Put(resultFilePath, storagePath)
		if err != nil {
			data.Job.Status = Failed
			log.Printf("Processing job %d  error %s\n", data.Job.ID, err.Error())
			details = err.Error()
			err = os.Remove(tempPath)
			if err != nil {
				details += ";" + err.Error()
			}
		}
		results = append(results, &JobResult{
			FormatId: format.ID,
			JobId:    data.Job.ID,
			Resource: publicUrl,
			Status:   data.Job.Status,
			Details:  details,
		})
		if data.Job.Status == Failed {
			break
		}
	}

	w.jobStorage.createJobResults(results)
	w.jobStorage.update(&data.Job)

	err = os.Remove(tempPath)
	if err != nil {
		return
	}
	log.Println("Processing completed")
}

func (w *Worker) generateRandomFilePath() string {
	return w.tempMediaPath + "/" + uuid.New().String()
}
func (w *Worker) fetchSource(url string) (string, error) {
	path := w.generateRandomFilePath()
	out, err := os.Create(path)
	if err != nil {
		return "", err
	}
	defer out.Close()

	log.Println("Fetching data from url " + url)
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	log.Println(out, resp.Status)

	_, err = io.Copy(out, resp.Body)
	if err != nil {
		return "", err
	}
	log.Println("Source fetched to " + path)
	return path, nil
}
