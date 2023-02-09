package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Edbeer/paymentapi/api"
	"github.com/Edbeer/paymentapi/config"
	"github.com/Edbeer/paymentapi/storage/psql"
	"github.com/Edbeer/paymentapi/storage/redis"

	"github.com/Edbeer/paymentapi/pkg/db/psql"
	"github.com/Edbeer/paymentapi/pkg/db/redis"
)

// @title           Payment Application
// @version         1.0
// @description     Simple payment system

// @securitydefinitions.apikey
// @in header
// @name Authorization
func main() {
	// init config
	config := config.GetConfig()

	// init postgres
	db, err := psql.NewPostgresDB(config)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("init postgres")

	// init redis
	redisClient := red.NewRedisClient()
	defer redisClient.Close()
	log.Println("init redis")

	psql := postgres.NewPostgresStorage(db)
	redisStore := redisrepo.NewRedisStorage(redisClient)
	// init server
	log.Println("init server")
	s := api.NewJSONApiServer(config, db, redisClient, psql, redisStore)
	go func() {
		s.Run()
	}()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<- quitCh

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	s.Server.Shutdown(ctx)
}