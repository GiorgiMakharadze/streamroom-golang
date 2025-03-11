package websocket

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/GiorgiMakharadze/video-stream-processor-golang.git/internal/rooms"
	pkg "github.com/GiorgiMakharadze/video-stream-processor-golang.git/pkg"
	"github.com/gorilla/websocket"
)

type AuthResponse struct {
	Authenticated bool `json:"authenticated"`
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  65536,
	WriteBufferSize: 65536,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type Response struct {
	StreamKey string `json:"streamKey,omitempty"`
	HLSURL    string `json:"hlsUrl,omitempty"`
	Message   string `json:"message,omitempty"`
}

func HandleStreamsList(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodOptions {
		w.WriteHeader(http.StatusNoContent)
		return
	}

	type StreamInfo struct {
		StreamKey string    `json:"streamKey"`
		CreatedAt time.Time `json:"createdAt"`
	}

	roomsList := rooms.Manager.ListRooms()

	var streams []StreamInfo
	for _, room := range roomsList {
		streams = append(streams, StreamInfo{
			StreamKey: room.ID,
			CreatedAt: room.CreatedAt,
		})
	}

	response := map[string]interface{}{
		"count":   len(streams),
		"streams": streams,
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Printf("Error encoding streams list: %v", err)
		http.Error(w, "failed to fetch streams list", http.StatusInternalServerError)
	}
}

func WsHandler(w http.ResponseWriter, r *http.Request, cfg *pkg.Config) {
	role := r.URL.Query().Get("role")
	switch role {
	case "publisher":
		handlePublisher(w, r, cfg)
	case "viewer":
		handleViewer(w, r, cfg)
	default:
		http.Error(w, "role query parameter required (publisher/viewer)", http.StatusBadRequest)
	}
}

func handlePublisher(w http.ResponseWriter, r *http.Request, cfg *pkg.Config) {
	streamKey := r.URL.Query().Get("streamKey")
	if streamKey == "" {
		http.Error(w, "streamKey query parameter required", http.StatusBadRequest)
		return
	}

	cookie := r.Header.Get("Cookie")
	if cookie == "" {
		http.Error(w, "Unauthorized: Missing session cookie", http.StatusUnauthorized)
		return
	}

	authenticated, err := validateUserSession(cookie, cfg)
	if err != nil || !authenticated {
		http.Error(w, "Unauthorized: Invalid session", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		http.Error(w, "failed to upgrade connection", http.StatusInternalServerError)
		return
	}
	log.Printf("Publisher connected with streamKey: %s", streamKey)

	room, err := rooms.Manager.CreateRoomWithKey(streamKey, conn)
	if err != nil {
		log.Printf("Failed to create room: %v", err)
		http.Error(w, err.Error(), http.StatusConflict)
		conn.Close()
		return
	}

	if err := room.StartFFmpeg(cfg); err != nil {
		log.Printf("Failed to start FFmpeg for streamKey %s: %v", streamKey, err)
		conn.Close()
		rooms.Manager.DeleteRoom(streamKey)
		return
	}

	go room.Monitor()

	resp := Response{
		StreamKey: streamKey,
		Message:   "Room created",
	}
	if err := conn.WriteJSON(resp); err != nil {
		log.Println("Error sending streamKey to publisher:", err)
		conn.Close()
		rooms.Manager.DeleteRoom(streamKey)
		return
	}

	for {
		messageType, data, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error (streamKey: %s): %v", streamKey, err)
			break
		}
		if messageType == websocket.BinaryMessage {
			select {
			case room.DataChan <- data:
			default:
				log.Printf("Room %s data channel full - dropping packet", streamKey)
			}
		}
	}

	close(room.DataChan)
	conn.Close()

	log.Printf("Stream %s closed", streamKey)
	rooms.Manager.DeleteRoom(streamKey)
}

func handleViewer(w http.ResponseWriter, r *http.Request, cfg *pkg.Config) {
	streamKey := r.URL.Query().Get("streamKey")
	if streamKey == "" {
		http.Error(w, "streamKey query parameter required", http.StatusBadRequest)
		return
	}

	_, exists := rooms.Manager.GetRoom(streamKey)
	if !exists {
		http.Error(w, "stream not found", http.StatusNotFound)
		return
	}

	hlsURL := fmt.Sprintf("%s/%s.m3u8", cfg.HLSBaseURL, streamKey)
	resp := Response{
		HLSURL:  hlsURL,
		Message: "Stream found",
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		log.Printf("Error responding to viewer request (streamKey: %s): %v", streamKey, err)
	}
}

func validateUserSession(cookie string, cfg *pkg.Config) (bool, error) {
	req, err := http.NewRequest("GET", cfg.AuthCheckURL, nil)
	if err != nil {
		return false, err
	}

	req.Header.Set("Cookie", cookie)

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return false, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return false, fmt.Errorf("non-200 response: %d", resp.StatusCode)
	}

	var authResp AuthResponse
	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return false, err
	}

	return authResp.Authenticated, nil
}
