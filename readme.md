# Person Detection - Python & Go

## Deskripsi Projek

Projek ini adalah sistem deteksi orang (person detection) berbasis model TensorFlow Lite yang dijalankan menggunakan Python, dengan backend server Go untuk pengambilan gambar dari stream (RTSP atau webcam), menjalankan deteksi secara periodik, dan mengirimkan hasil deteksi ke client melalui WebSocket.

Fitur utama:

- Mengambil frame dari stream RTSP atau webcam menggunakan FFmpeg.
- Deteksi objek "person" pada gambar menggunakan model TFLite.
- Menyimpan gambar hasil deteksi dengan bounding box (maksimal 2 gambar per deteksi).
- Backend Go untuk otomatisasi, logging, dan komunikasi WebSocket.

---

## Struktur Folder

```
person-detection/
├── detection/
│   ├── detect.py
│   └── model/
│       ├── 1.tflite
│       └── labelmap.txt
├── main.go
```

---

## Prasyarat

- **Python 3.x**
- **Go** (disarankan versi terbaru)
- **FFmpeg** (pastikan sudah di-`PATH`)
- **pip** untuk instalasi dependensi Python

### Library Python yang dibutuhkan

- opencv-python
- numpy
- tensorflow

Install dengan:

```sh
pip install opencv-python numpy tensorflow
```

---

## Langkah Menjalankan Projek

### 1. Clone atau Download Projek

```sh
git clone https://github.com/Sahlil23/person-detection.git
cd person-detection
```

### 2. Pastikan FFmpeg sudah terinstall

Cek dengan:

```sh
ffmpeg -version
```

Jika belum ada, download dari [https://ffmpeg.org/download.html](https://ffmpeg.org/download.html) dan tambahkan ke PATH.

### 3. Install Dependensi Python

```sh
pip install opencv-python numpy tensorflow
```

### 4. Jalankan Server Go

```sh
go run main.go
```

Server akan berjalan di `http://localhost:8080`.

### 5. Koneksi Client WebSocket

Client dapat terhubung ke endpoint:

```
ws://localhost:8080/ws
```

dan akan menerima hasil deteksi dalam format JSON.

### 6. Hasil Deteksi

- Gambar hasil deteksi dengan bounding box akan otomatis dibuat dengan nama:
  - `detection_result.jpg`
- Hasil deteksi juga dikirim ke client WebSocket.

---

## Pengaturan Sumber Video

Secara default, projek menggunakan RTSP stream.  
Untuk menggunakan webcam, ubah kode di `main.go` pada bagian FFmpeg command menjadi seperti berikut:

```go
cmdFFmpeg := exec.Command(
    "ffmpeg",
    "-f", "dshow",
    "-i", "video=Integrated Camera", // atau nama device webcam Anda
    "-vframes", "1",
    "-q:v", "2",
    "-y", framePath,
)
```

Pastikan nama device webcam sesuai dengan yang ada di komputer Anda.

---

## Troubleshooting

- Jika deteksi tidak berjalan, cek log di terminal Go dan Python.
- Pastikan model TFLite dan labelmap.txt sudah ada di folder `detection/model/`.
- Pastikan FFmpeg dapat mengambil gambar dari sumber video Anda.

---

## Lisensi

Projek ini bebas digunakan untuk pembelajaran dan pengembangan lebih
