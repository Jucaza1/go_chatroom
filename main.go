package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
)

func main() {
	//create hub
	ctx,cancel := context.WithCancel(context.Background())
    defer cancel()
	hub := NewHub(ctx)
	//go hub.Run()
	go hub.Run()
	//create mux
	mux := http.NewServeMux()
	//add websocket handler
	mux.HandleFunc("GET /join", MakeWSHandler(ctx, &hub))
	//add page client handler
	mux.HandleFunc("GET /", func(w http.ResponseWriter, r *http.Request) {
		log.Println(r.URL)
		if r.URL.Path != "/" {
			http.Error(w, "Not found", http.StatusNotFound)
			return
		}
		// if r.Method != http.MethodGet {
		// 	http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		// 	return
		// }
		http.ServeFile(w, r, "home.html")
	})
	//listen and serve
    fmt.Println("server listening on: http://localhost:3000")
    http.ListenAndServe(":3000",mux)
    fmt.Println("server listening on: http://localhost:3000")
}
