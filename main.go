package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Edbeer/paymentapi/api"
	"github.com/Edbeer/paymentapi/storage/psql"
	"github.com/Edbeer/paymentapi/storage/redis"

	"github.com/Edbeer/paymentapi/pkg/db/psql"
	"github.com/Edbeer/paymentapi/pkg/db/redis"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := psql.NewPostgresDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	log.Println("init postgres")

	redisClient := red.NewRedisClient()
	defer redisClient.Close()
	log.Println("init redis")

	psql := postgres.NewPostgresStorage(db)
	redisStore := redisrepo.NewRedisStorage(redisClient)
	log.Println("init server")
	s := api.NewJSONApiServer(":8080", db, redisClient, psql, redisStore)
	go func() {
		s.Run()
	}()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case signal := <- quitCh:
		log.Println(signal)
	}

	s.Server.Shutdown(ctx)
}