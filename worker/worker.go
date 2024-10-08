package worker

import (
	"context"
	"sync"
	"time"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/messaging"
	"github.com/go-co-op/gocron"
	"github.com/Roofiif/sim-graha-nirmala-worker/config"
	"github.com/Roofiif/sim-graha-nirmala-worker/db"
	"github.com/Roofiif/sim-graha-nirmala-worker/logger"
	"google.golang.org/api/option"
)

type worker struct {
	s         *gocron.Scheduler
	db        *db.Client
	fcmClient *messaging.Client
	mu        sync.Mutex
	checked   map[string]bool
	lastNotificationDate map[int64]time.Time
}

func NewWorker(s *gocron.Scheduler, db *db.Client) *worker {
	cfg := config.Cfg()

	credentialsJSON := cfg.Fr.Credentials
	if credentialsJSON == "" {
		logger.Log().Fatal("missing GOOGLE_CREDENTIALS in config file")
	}

	opt := option.WithCredentialsJSON([]byte(credentialsJSON))
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		logger.Log().Fatal("error initializing app", "error", err)
	}

	fcmClient, err := app.Messaging(context.Background())
	if err != nil {
		logger.Log().Fatal("error getting Messaging client", "error", err)
	}

	return &worker{
		s:         s,
		db:        db,
		fcmClient: fcmClient,
		checked:   make(map[string]bool),
		lastNotificationDate: make(map[int64]time.Time),
	}
}

func (w *worker) Do() {
	w.s.Every(5).Seconds().Do(w.generateMonthlyBilling)
	w.s.Every(1).Day().Do(w.send15thDayNotification)
	w.s.Every(1).Day().Do(w.send22thDayNotification)
	
	logger.Log().Info("worker started")
	w.s.StartAsync()
	logger.Log().Info("Tasks executed")

	select{}
}
