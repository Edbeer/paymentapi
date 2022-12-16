package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Edbeer/paymentapi/internal/handlers"
	"github.com/Edbeer/paymentapi/internal/storage"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := storage.NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}
	log.Println("init server")
	s := handlers.NewJSONApiServer(":8080", db)
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