package main

func main() {
	s := NewJSONApiServer(":8080")
	s.Run()
}