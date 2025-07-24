package main

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/callumj/clicky/pkg/cameras"
	"github.com/callumj/clicky/pkg/config"
	"github.com/callumj/clicky/pkg/storage"
	"github.com/callumj/clicky/pkg/storage/local"
	"github.com/callumj/clicky/pkg/storage/multi"
	"github.com/callumj/clicky/pkg/storage/s3"

	"github.com/rs/zerolog/log"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"

	"github.com/go-co-op/gocron/v2"
)

func main() {
	// load config from file specified in arguments
	configPath := os.Args[1]
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		fmt.Printf("Error loading config: %v\n", err)
		return
	}

	// initialize storage based on config
	var storage storage.Storage
	if cfg.Storage != nil {
		mStore := multi.NewMultiStorage()
		if cfg.Storage.S3 != nil {
			s3Store, err := s3.NewS3Storage(cfg.Storage.S3)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to initialize S3 storage")
				return
			}
			mStore.Register("s3", s3Store)
		}

		if cfg.Storage.Local != nil {
			mStore.Register("local", local.NewLocalStorage(cfg.Storage.Local))
		}
		storage = mStore
	}

	if storage == nil {
		log.Fatal().Msg("No storage configured")
		return
	}

	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}

	// initialize snapshotter
	snapshotter := cameras.NewSnapshotterWithClient(client, cfg, storage)

	// create scheduler
	var loc *time.Location
	if cfg.Tz != "" {
		loc, err = time.LoadLocation(cfg.Tz)
		if err != nil {
			log.Fatal().Err(err).Msgf("Failed to load timezone location")
			return
		}
	} else {
		loc = time.Local // default to local timezone
	}
	s, err := gocron.NewScheduler(gocron.WithLocation(loc))
	if err != nil {
		log.Fatal().Err(err).Msg("Error creating scheduler")
		return
	}
	log.Info().Str("location", loc.String()).Msg("Scheduler created")

	j, err := s.NewJob(
		gocron.CronJob(cfg.Snapshot.CronSchedule, false),
		gocron.NewTask(
			func() {
				if err := snapshotter.SaveSnapshots(); err != nil {
					log.Error().Err(err).Msg("Error saving snapshots")
				}
			},
		),
	)

	if err != nil {
		log.Fatal().Err(err).Msg("Error creating job")
		return
	}

	nextRun, err := j.NextRun()
	if err != nil {
		log.Fatal().Err(err).Msg("Error getting next run time")
		return
	}
	log.Info().Time("nextRun", nextRun).Msgf("Cron scheduler started")
	s.Start()

	// initialize web server
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	// attach routes
	e.POST("/snapshot", func(c echo.Context) error {
		if err := snapshotter.SaveSnapshots(); err != nil {
			return c.JSON(http.StatusInternalServerError, map[string]string{"error": err.Error()})
		}
		return c.JSON(http.StatusOK, map[string]string{"status": "snapshots saved"})
	})

	go func() {
		if err := e.Start(":8080"); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatal().Err(err).Msg("failed to start server")
			return
		}
	}()

	log.Info().Msg("Server started on :8080")
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt)
	<-done

	log.Info().Msg("Shutting down...")
	err = s.Shutdown()
	if err != nil {
		log.Fatal().Err(err).Msg("Error shutting down scheduler")
	} else {
		log.Info().Msg("Scheduler shut down successfully")
	}

	err = e.Shutdown(context.TODO())
	if err != nil {
		log.Fatal().Err(err).Msg("Error shutting down server")
	} else {
		log.Info().Msg("Server shut down successfully")
	}
}
