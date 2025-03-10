# Streamroom-golang

This project is a **WebSocket-based video processing server** built with **Go**. It allows clients to send video streams in chunks over WebSocket, processes the received video using **FFmpeg**, converts it to FLV format, and then streams the output to an **RTMP server (Nginx-RTMP)** for further handling (e.g., live streaming or storage).

## How It Works

1. **WebSocket Connection:**
   - Clients connect to the WebSocket server (`/ws` endpoint).
   - They send video data in binary chunks over the connection.
   
2. **Receiving Video Data:**
   - The server collects these chunks and writes them into a temporary file.
   
3. **Processing Video:**
   - Once the full video is received, a task is queued for processing.
   - The task is picked up by a worker from a pool of concurrent workers.
   - FFmpeg converts the received file into FLV format.
Here’s the full `README.md` content in proper `.md` format for you to copy-paste directly:

```md
# Video Stream Processor

This project is a **WebSocket-based real-time video streaming server** built with **Go**. It allows clients (like browsers) to send **live video streams** over WebSockets, and the server forwards the stream to an **RTMP server (Nginx-RTMP)** using **FFmpeg**. The RTMP server then **transmuxes** the stream into HLS segments for playback.

---

## 🛠️ How It Works

### 1️⃣ WebSocket Connection (Publisher Role)
- Clients connect to the WebSocket server at `/ws?role=publisher&streamKey=<key>`.
- Clients send **video data in binary chunks** over the WebSocket connection.

### 2️⃣ Room Management
- Each stream (based on `streamKey`) creates a **Room** (a managed stream session).
- Each Room gets its own **goroutine** to handle:
    - Receiving data from WebSocket.
    - Sending data into an FFmpeg process via `stdin`.
    - Monitoring FFmpeg’s health and lifecycle.

### 3️⃣ Real-time Processing (FFmpeg)
- FFmpeg receives data **directly from the WebSocket via a Go channel**.
- FFmpeg transcodes the incoming data into **FLV format** and pushes it to the configured RTMP URL:
```
rtmp://<nginx-rtmp>/live/<streamKey>
```

---

### 4️⃣ RTMP to HLS (Nginx-RTMP)
- Nginx-RTMP receives the stream and **transmuxes** it into HLS segments.
- HLS playlist and segments are served via:
```
http://localhost:8080/hls/live/<streamKey>.m3u8
```

---

### 5️⃣ WebSocket Connection (Viewer Role)
- Viewers can request the list of active streams via:
```
GET /streams
```
- Viewers can get the playback URL for a given stream:
```
http://localhost:8080/hls/live/<streamKey>.m3u8
```

---

## ⚡ Features

**WebSocket support** for real-time video ingest (publishers).  
**Live stream management** using **Goroutines and Channels**.  
**FFmpeg integration** for real-time transcoding (WebM to FLV).  
**RTMP streaming support** for live broadcasting.  
**HLS playback** via Nginx-RTMP.  
**Graceful cleanup** — when a stream ends, FFmpeg is stopped and HLS files are removed.  
**Health monitoring** of the FFmpeg process (detect crashes and recover).

---

## 🌐 Architecture Diagram

```text
[ Publisher ] --(WebSocket)--> [ Go Server ] --(FLV over RTMP)--> [ Nginx-RTMP ]
                         ^                                            |
                         |                                            v
                    [ Viewer ] <----(HLS over HTTP)---- [ Nginx HLS Server ]
```

---

## ⚙️ Environment Variables

| Variable | Description | Default |
|---|---|---|
| `WS_PORT` | WebSocket server port | `9090` |
| `RTMP_BASE_URL` | RTMP server base URL | `rtmp://localhost/live` |
| `HLS_BASE_URL` | HLS playback base URL | `http://localhost:8080/hls/live` |

---

## 🚀 Usage

### 1. Start the Backend
```bash
go run main.go
```

### 2. Start the Nginx-RTMP Container
```bash
docker-compose up nginx-rtmp
```

---

### 3. Publish a Stream (Publisher)
A browser can use:
```
ws://localhost:9090/ws?role=publisher&streamKey=myAwesomeStream
```

---

### 4. Watch the Stream (Viewer)
```
http://localhost:8080/hls/live/myAwesomeStream.m3u8
```

---

## 📦 Dependencies

| Library/Tool | Purpose |
|---|---|
| `gorilla/websocket` | WebSocket handling |
| `FFmpeg` | Video processing and transcoding |
| `Nginx-RTMP` | RTMP ingest + HLS output |

---

## Advantages After Refactor

No more blocking — every stream runs in its own goroutine.  
Efficient `Room` lifecycle — clean start, monitoring, and cleanup.  
Live logs from FFmpeg directly into the server logs for easy debugging.  
No need to buffer full files — direct piping from WebSocket to FFmpeg.

---

## Important Note

- To reduce stream startup time, you can tune:
```nginx
hls_fragment 500ms;
hls_playlist_length 1s;
```
- For ultra-low-latency HLS, additional nginx-rtmp patches may be required (optional).

---

## 🎉 Conclusion

This system ensures **efficient, scalable, real-time video processing** with Go and modern streaming technologies like RTMP + HLS, while keeping resource usage optimized using **channels and goroutines**.
```

---

Let me know if you want me to generate a **ready-to-use `README.md` file** and package it into a zip with your `rooms.go`, `handler.go`, `config.go`, and `nginx.conf`.  
Want me to do that? ✅
4. **Streaming to RTMP Server:**
   - The converted FLV file is streamed to an RTMP endpoint (configured in environment variables).
   - After streaming, the temporary files are deleted to free up space.

## Features
- **WebSocket support** for real-time video streaming.
- **Asynchronous processing** using a worker pool to handle multiple uploads efficiently.
- **FFmpeg integration** for video conversion to FLV format.
- **RTMP streaming support** for live video broadcasting.
- **Optimized resource management** with temporary file cleanup and background processing.

## Environment Variables
The server reads configurations from environment variables:
- `WS_PORT`: WebSocket server port (default: `9090`)
- `RTMP_URL`: RTMP server URL (default: `rtmp://localhost/live/stream`)
- `HLS_URL`: HLS playback URL (default: `http://localhost:8080/hls/live/stream.m3u8`)

## Usage
1. Start the Go WebSocket server.
2. Connect a client and start sending a video stream.
3. The server will process the video asynchronously and stream it to the RTMP server.

## Dependencies
- `gorilla/websocket` (WebSocket handling)
- `FFmpeg` (Video processing and conversion)
- `Nginx-RTMP` (Streaming server)

This system ensures **efficient, concurrent, and scalable** video processing while keeping the WebSocket server responsive.

