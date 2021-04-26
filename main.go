package main

import (
	"log"
	"os"
	"os/signal"
	"preemptible-lifecycle-scheduler/cluster"
	"preemptible-lifecycle-scheduler/config"
	"preemptible-lifecycle-scheduler/peakhour"
	"preemptible-lifecycle-scheduler/scheduler"
	"sync"
	"syscall"
)

func main() {
	cfg := config.NewDefaultConfig()
	err := cfg.Load("config.yaml")
	if err != nil {
		log.Fatalf("failed to read config file: %v", err)
	}
	log.Printf("using configuration: %#v", cfg)

	ph := peakhour.NewPeakHour(cfg.PeakHourRanges)

	clusterClient, err := cluster.NewClient(cfg)
	if err != nil {
		log.Fatalf("failed to init kubernetes client: %v", err)
	}

	schedulerClient := scheduler.NewClient(clusterClient, ph)

	gracefulShutdown := make(chan os.Signal)
	signal.Notify(gracefulShutdown, syscall.SIGTERM, syscall.SIGINT)
	waitGroup := &sync.WaitGroup{}

	go func(waitGroup *sync.WaitGroup) {
		schedulerClient.Start()
	}(waitGroup)

	signalReceived := <-gracefulShutdown
	log.Printf("received signal %v", signalReceived)
	waitGroup.Wait()
	log.Printf("shutting down...")
}
