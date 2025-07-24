package main

import (
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"time"

	"github.com/callumj/clicky/pkg/cameras"
	"github.com/callumj/clicky/pkg/config"
	"github.com/callumj/clicky/pkg/storage"
	"github.com/callumj/clicky/pkg/storage/local"
	"github.com/callumj/clicky/pkg/storage/s3"

	"github.com/rs/zerolog/log"

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
		if cfg.Storage.S3 != nil {
			storage, err = s3.NewS3Storage(cfg.Storage.S3)
			if err != nil {
				log.Fatal().Err(err).Msg("Failed to initialize S3 storage")
				return
			}
		} else if cfg.Storage.Local != nil {
			storage = local.NewLocalStorage(cfg.Storage.Local)
		}
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

	loc := time.Local
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
}
