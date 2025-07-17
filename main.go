package main

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"os/exec"
	"sync"
	"time"
	"bytes"
	"github.com/gorilla/websocket"
)

// URL Stream RTSP Anda (ganti dengan URL asli)
const rtspURL = "rtsp://username:pw@1ip:port/Channel" // Ganti dengan URL RTSP yang sesuai

// Konfigurasi WebSocket
var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Izinkan koneksi dari mana saja (untuk development)
	},
}

// Pool untuk menyimpan koneksi klien WebSocket
var clients = make(map[*websocket.Conn]bool)
var clientsMutex = sync.Mutex{}

// Struct untuk menampung hasil deteksi dari Python
type DetectionResult struct {
	Persons []struct {
		Box   []int   `json:"box"`
		Score float64 `json:"score"`
	} `json:"persons"`
}

// Fungsi untuk menangani koneksi WebSocket
func handleConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Fatal(err)
	}
	defer ws.Close()

	clientsMutex.Lock()
	clients[ws] = true
	clientsMutex.Unlock()

	log.Println("Client baru terhubung")

	// Loop agar koneksi tetap terbuka, juga untuk menghapus klien saat koneksi terputus
	for {
		_, _, err := ws.ReadMessage()
		if err != nil {
			clientsMutex.Lock()
			delete(clients, ws)
			clientsMutex.Unlock()
			log.Println("Client terputus")
			break
		}
	}
}

// Fungsi untuk menyiarkan pesan ke semua klien yang terhubung
func broadcastMessage(message []byte) {
	clientsMutex.Lock()
	defer clientsMutex.Unlock()
	for client := range clients {
		err := client.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			log.Printf("error: %v", err)
			client.Close()
			delete(clients, client)
		}
	}
}

// Fungsi utama yang melakukan deteksi secara periodik
func startDetectionLoop() {
	ticker := time.NewTicker(2 * time.Second) // Lakukan deteksi setiap 2 detik
	defer ticker.Stop()

	for range ticker.C {
		log.Println("Mengambil frame dari stream...")

		// 1. Gunakan FFmpeg untuk mengambil satu frame
		framePath := "temp_frame.jpg"
		cmdFFmpeg := exec.Command("ffmpeg", "-i", rtspURL, "-vframes", "1", "-q:v", "2", "-y", framePath)
		if err := cmdFFmpeg.Run(); err != nil {
			log.Printf("Gagal mengambil frame: %v", err)
			continue
		}

		// 2. Jalankan skrip Python untuk deteksi
		log.Println("Menjalankan deteksi...")
		cmdPython := exec.Command("py", "detection/detect.py", framePath)

		var stderr bytes.Buffer // Siapkan buffer untuk menampung error
		cmdPython.Stderr = &stderr

		output, err := cmdPython.Output() // Hanya mengambil stdout
		if err != nil {
			// Jika ada error, cetak pesan error dari stderr untuk debugging
			log.Printf("Gagal menjalankan skrip Python: %v, Stderr: %s", err, stderr.String())
			continue
		}

		// 3. Hapus file frame sementara
		os.Remove(framePath)

		// 4. Pastikan output JSON valid sebelum di-broadcast
		var result DetectionResult
		if err := json.Unmarshal(output, &result); err != nil {
			log.Printf("Output JSON tidak valid dari Python: %v", err)
			continue
		}
		
		if len(result.Persons) > 0 {
			// KASUS 1: Orang ditemukan
			log.Printf("✅ Deteksi Berhasil: %d orang ditemukan.", len(result.Persons))
			broadcastMessage(output)
		} else {
			// KASUS 2: Tidak ada orang yang ditemukan
			log.Println("✔️ Deteksi Berhasil: Tidak ada orang yang ditemukan.")
		}
	}
}

func main() {
	// Jalankan loop deteksi di goroutine terpisah
	go startDetectionLoop()

	// Setup server HTTP untuk WebSocket
	http.HandleFunc("/ws", handleConnections)
	
	// Server juga bisa menyajikan stream HLS jika dikonfigurasi
	// http.Handle("/", http.FileServer(http.Dir("./hls")))

	log.Println("Server berjalan di http://localhost:8080")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		log.Fatal("ListenAndServe: ", err)
	}
}