package main

import (
	"log"
	"net/http"

	"github.com/GiorgiMakharadze/video-stream-processor-golang.git/internal/middleware"
	"github.com/GiorgiMakharadze/video-stream-processor-golang.git/internal/websocket"
	pkg "github.com/GiorgiMakharadze/video-stream-processor-golang.git/pkg"
)

func main() {
	cfg := pkg.LoadConfig()

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("ok"))
	})
	http.HandleFunc("/streams", middleware.WithCORS(websocket.HandleStreamsList))
	http.HandleFunc("/ws", func(w http.ResponseWriter, r *http.Request) {
		websocket.WsHandler(w, r, cfg)
	})

	port := ":" + cfg.WebSocketPort
	log.Println("Server is running on:", port)
	if err := http.ListenAndServe(port, nil); err != nil {
		log.Fatal("Server error:", err)
	}
}
