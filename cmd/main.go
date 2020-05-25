package main

import (
	"context"
	graylog "github.com/gemnasium/logrus-graylog-hook"
	log "github.com/sirupsen/logrus"
	"github.com/riceChuang/tahmkench/config"
	"github.com/riceChuang/tahmkench/db"
	"github.com/riceChuang/tahmkench/service"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	//init config
	appConfig := config.InitialAppConfig()
	agentConfig := config.InitialAgentConfig()
	err := db.SetStorage(appConfig)
	if err != nil {
		log.Panic(err)
	}
	//set log
	logLevel, err := log.ParseLevel(appConfig.Logs.LogLevel)
	if err != nil {
		log.Warnf("log level invalid, set log to info level, config data:%s", appConfig.Logs.LogLevel)
		logLevel = log.InfoLevel
	}
	log.SetLevel(logLevel)
	//graylog hook
	gLog := graylog.NewGraylogHook(appConfig.Logs.Addr, map[string]interface{}{"this": "is logged every time"})
	gLog.Level = log.GetLevel()
	log.AddHook(gLog)

	//initial manager and record queue
	manager := service.NewWarehouseManager()
	//init and start collection source
	err = manager.InitialSource(agentConfig.Sources)
	if err != nil {
		log.Panic(err)
	}
	manager.RunSource()
	//init and start sender record
	manager.InitialAndMonitorMatch(agentConfig.Matches)
	//pop collection queue and dispatch to sender
	manager.PopSourceRecord(&manager.WarehouseQueue, appConfig.ManagerWorker)

	srv := &http.Server{
		Addr: appConfig.BindPort,
	}
	go func() {
		// service connections
		log.Info("tahmkench Started")
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// waiting signal and graceful shut down（5 sec over time)
	quit := make(chan os.Signal)
	// syscall SIGINT:ctrl-c, SIGTSTP:ctrl-z, SIGQUIT:ctrl-\
	signal.Notify(quit, os.Interrupt, syscall.SIGQUIT, syscall.SIGINT, syscall.SIGTSTP)
	<-quit
	log.Info("Shutdown tahmkench ...")

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server Shutdown:", err)
	}
	log.Println("Server tahmkench exiting")

}
