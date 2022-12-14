package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Edbeer/paymentapi/api"
	"github.com/Edbeer/paymentapi/storage"
	"github.com/Edbeer/paymentapi/pkg/db"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := database.NewPostgresDB()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()
	
	psql := storage.NewPostgresStorage(db)

	log.Println("init server")
	s := api.NewJSONApiServer(":8080", db, psql)
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