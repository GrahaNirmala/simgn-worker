package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/go-co-op/gocron"
	"github.com/Roofiif/sim-graha-nirmala-worker/db"
	"github.com/Roofiif/sim-graha-nirmala-worker/logger"
	"github.com/Roofiif/sim-graha-nirmala-worker/worker"
)

func main() {
	loc, err := time.LoadLocation("Asia/Jakarta")
    if err != nil {
        logger.Log().Fatal("Failed to load location:", err)
    }

    s := gocron.NewScheduler(loc)
	dbClient, err := db.NewClient()
	if err != nil {
		logger.Log().Fatal(err.Error())
	}
	defer dbClient.Close()

	w := worker.NewWorker(s, dbClient)
	http.HandleFunc("/run-tasks", func(wr http.ResponseWriter, r *http.Request) {
		w.Do() // This starts the scheduled tasks
		fmt.Fprintln(wr, "Tasks executed")
	})

	// Start the HTTP server
	logger.Log().Info("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		logger.Log().Fatal("Failed to start server:", err)
	}
}
