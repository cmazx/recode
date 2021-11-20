package main

import (
	"fmt"
	"github.com/cmazx/recode/pkg/queue"
	"github.com/cmazx/recode/pkg/recode"
	"github.com/cmazx/recode/pkg/storage"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"os"
	"os/signal"
	"syscall"
)

func newDbConnection() *gorm.DB {
	conn, err := gorm.Open(postgres.Open(os.Getenv("DB_DSN")), &gorm.Config{})
	if err != nil {
		_, err := fmt.Fprintf(os.Stderr, "Unable to connect to database: %v\n", err)
		if err != nil {
			//nothing
		}
		os.Exit(1)
	}

	err = recode.AutomigrateJob(conn)
	if err != nil {
		panic(err)
	}
	err = recode.AutoMigrateFormats(conn)
	if err != nil {
		panic(err)
	}

	return conn
}

func main() {
	var sigs = make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
	if len(os.Args) < 2 {
		panic("No role argument specified")
	}

	db := newDbConnection()
	formatStorage := recode.NewMediaFormatStorage(db)
	jobStorage := recode.NewJobFormatStorage(db)
	stg := storage.NewS3Storage()

	switch os.Args[1] {
	case "api":
		producer, err := queue.NewProducer()
		if err != nil {
			panic(err)
		}
		recode.NewWorkBroker(producer, formatStorage, jobStorage, sigs).Start()
	case "worker":
		consumer, err := queue.NewConsumer()
		if err != nil {
			panic(err)
		}
		recode.NewWorker(consumer, formatStorage, jobStorage, stg, sigs).Start()
	default:
		panic("Unknown role " + os.Args[1] + " specified")
	}
}
