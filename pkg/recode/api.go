package recode

import (
	"fmt"
	"github.com/cmazx/recode/pkg/queue"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"log"
	"net/http"
	"os"
	"strconv"
)

type ApiService struct {
	jobs          chan Job
	signalCh      chan os.Signal
	queueService  *queue.Queue
	formatStorage *MediaFormatStorage
	jobStorage    *JobStorage
}

func NewWorkBroker(queueService *queue.Queue, formatStorage *MediaFormatStorage, storage *JobStorage, signalCh chan os.Signal) *ApiService {
	return &ApiService{
		make(chan Job, 1000),
		signalCh,
		queueService,
		formatStorage,
		storage,
	}
}

func (w *ApiService) Start() {
	log.Println("Api server start")
	e := echo.New()
	e.Use(middleware.LoggerWithConfig(middleware.LoggerConfig{
		Format: "method=${method}, uri=${uri}, status=${status}\n",
	}))
	e.GET("formats", w.listFormats)
	e.POST("formats", w.CreateFormat)
	e.PUT("formats/:id", w.UpdateFormat)
	e.DELETE("formats/:id", w.DeleteFormat)

	e.GET("jobs", w.JobList)
	e.POST("jobs", w.CreateJob)
	e.DELETE("jobs/:id", w.DeleteJob)

	go func() {
		e.Logger.Fatal(e.Start(":" + os.Getenv("API_PORT")))
	}()

	<-w.signalCh
	log.Println("Api server stopped")
}

func (w *ApiService) CreateJob(c echo.Context) error {
	job := new(Job)
	err := c.Bind(job)
	if err != nil {
		return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
	}
	job.Status = New
	w.jobStorage.create(job)

	err = w.queueService.Enqueue(w.queueService.JobTopicName, JobData{
		*job,
		w.formatStorage.listByIds(job.Formats),
	})
	if err != nil {
		w.jobStorage.delete(job.ID)
		return echo.NewHTTPError(http.StatusInternalServerError, "Error on enqueue job: "+err.Error())
	}

	return c.JSON(http.StatusOK, job)
}

func (w *ApiService) DeleteJob(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}

	w.jobStorage.delete(uint(id))

	return c.NoContent(http.StatusNoContent)
}

func (w *ApiService) JobList(c echo.Context) error {
	pageStr := c.QueryParam("page")
	page := 0
	if pageStr != "" {
		pageInt, err := strconv.Atoi(pageStr)
		if err != nil {
			return fmt.Errorf("page parameter not integer")
		}
		page = pageInt
	}

	list := w.jobStorage.list(page)

	return c.JSON(http.StatusOK, list)
}

func (w *ApiService) CreateFormat(c echo.Context) error {
	format := &MediaFormat{}
	err := c.Bind(format)
	if err != nil {
		return err
	}
	w.formatStorage.create(format)

	return c.JSON(http.StatusOK, format)
}
func (w *ApiService) UpdateFormat(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	format := w.formatStorage.find(id)
	if format == nil {
		return echo.NewHTTPError(http.StatusNotFound, "format not found")
	}
	err = c.Bind(format)
	if err != nil {
		return err
	}
	w.formatStorage.update(format)

	return c.JSON(http.StatusOK, format)
}
func (w *ApiService) listFormats(c echo.Context) error {
	log.Println("list formats")
	formats := w.formatStorage.list()
	return c.JSON(http.StatusOK, formats)
}
func (w *ApiService) DeleteFormat(c echo.Context) error {
	id, err := strconv.Atoi(c.Param("id"))
	if err != nil {
		return err
	}
	w.formatStorage.delete(id)

	return c.NoContent(http.StatusNoContent)
}

type Response struct {
	data interface{}
	meta interface{}
}
