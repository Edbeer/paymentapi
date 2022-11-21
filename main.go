package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	db, err := NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.InitTable(); err != nil {
		log.Fatal(err)
	}
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
}