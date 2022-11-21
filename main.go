package main

import "log"

func main() {
	db, err := NewPostgresStorage()
	if err != nil {
		log.Fatal(err)
	}
	if err := db.InitTable(); err != nil {
		log.Fatal(err)
	}
	s := NewJSONApiServer(":8080", db)
	s.Run()
}