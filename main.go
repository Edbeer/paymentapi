package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	db, err := NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}
	// if err := db.InitTables(ctx); err != nil {
	// 	log.Fatal(err)
	// }
	s := NewJSONApiServer(":8080", db)
	go func() {
		s.Run()
	}()

	quitCh := make(chan os.Signal, 1)
	signal.Notify(quitCh, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	select {
	case signal := <- quitCh:
		log.Println(signal)
	}

	s.server.Shutdown(ctx)
}