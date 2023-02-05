package main

import (
	"fmt"
	"io"
	"net/http"
)

func pong(w http.ResponseWriter, r *http.Request) {
	fmt.Println("Ping sent, returning pong")
	io.WriteString(w, "Pong")
}

func main() {
	// Handle pong message
	http.HandleFunc("/ping", pong)
	fs := http.FileServer(http.Dir("../../assets/"))

	// server index.html for root path and stuff
	http.Handle("/",fs)

	err := http.ListenAndServe(":9090", nil)
	if err != nil {
		fmt.Println("Failed to start server", err)
		return
	}
}
