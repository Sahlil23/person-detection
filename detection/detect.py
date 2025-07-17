import os
os.environ['TF_CPP_MIN_LOG_LEVEL'] = '2'

import cv2
import json
import sys
import numpy as np
import tensorflow as tf

# --- KONFIGURASI ---
CONFIDENCE_THRESHOLD = 0.6
# --------------------

# Ambil path gambar dari argumen command line
image_path = sys.argv[1]

# Path ke model
model_path = 'detection/model/1.tflite'

# Muat model TFLite
try:
    interpreter = tf.lite.Interpreter(model_path=model_path)
except Exception as e:
    print(json.dumps({"error": f"Gagal memuat model: {e}"}))
    sys.exit(1)
    
interpreter.allocate_tensors()
input_details = interpreter.get_input_details()
output_details = interpreter.get_output_details()
height = input_details[0]['shape'][1]
width = input_details[0]['shape'][2]

# Baca gambar
image = cv2.imread(image_path)
if image is None:
    print(json.dumps({"error": f"Tidak bisa membaca gambar dari {image_path}"}))
    sys.exit(1)

# Buat salinan gambar asli untuk digambar
image_with_boxes = image.copy()

image_rgb = cv2.cvtColor(image, cv2.COLOR_BGR2RGB)
imH, imW, _ = image.shape
image_resized = cv2.resize(image_rgb, (width, height))
input_data = np.expand_dims(image_resized, axis=0)

# Lakukan deteksi
interpreter.set_tensor(input_details[0]['index'], input_data)
interpreter.invoke()

# Ambil hasil deteksi
boxes = interpreter.get_tensor(output_details[0]['index'])[0]
classes = interpreter.get_tensor(output_details[1]['index'])[0]
scores = interpreter.get_tensor(output_details[2]['index'])[0]

# Proses hasil dan cari "person"
detections = []
for i in range(len(scores)):
    # Cek apakah skor di atas ambang batas
    if scores[i] > CONFIDENCE_THRESHOLD:
        
        # Berdasarkan hasil debug visual, ID untuk 'person' adalah 0
        if int(classes[i]) == 0:
            ymin = int(max(1, (boxes[i][0] * imH)))
            xmin = int(max(1, (boxes[i][1] * imW)))
            ymax = int(min(imH, (boxes[i][2] * imH)))
            xmax = int(min(imW, (boxes[i][3] * imW)))
            
            detections.append({
                "box": [xmin, ymin, xmax, ymax],
                "score": float(scores[i])
            })
            
            # ==========================================================
            ## TAMBAHAN: Gambar Bounding Box dan Teks pada Gambar
            # ==========================================================
            # Gambar kotak
            cv2.rectangle(image_with_boxes, (xmin, ymin), (xmax, ymax), (10, 255, 0), 2)
            
            # Siapkan teks label
            label = f'Person: {scores[i]:.2f}'
            label_ymin = max(ymin - 15, 15)
            
            # Gambar latar belakang untuk teks
            cv2.putText(image_with_boxes, label, (xmin, label_ymin), cv2.FONT_HERSHEY_SIMPLEX, 0.7, (10, 255, 0), 2)
            # ==========================================================

# ==========================================================
## TAMBAHAN: Simpan Gambar Hasil Deteksi
# ==========================================================
if len(detections) > 0:
    cv2.imwrite('detection_result.jpg', image_with_boxes)
# ==========================================================


# Cetak HANYA output JSON untuk Go
print(json.dumps({"persons": detections}))