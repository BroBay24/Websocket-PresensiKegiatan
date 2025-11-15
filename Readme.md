# README4 — Petunjuk Singkat & Demo Realtime

## Nama kelompok & Nim
christian B. Ferdinand / 233406006
Laurensius Ryan Antony / 233401014

## Deskripsi Singkat
Aplikasi "Sistem Presensi Kegiatan Kampus Berbasis QR Code dan WebSocket Realtime" adalah aplikasi untuk mencatat kehadiran peserta acara kampus. Peserta mengisi form (Nama, NIM, Jurusan, Angkatan) lewat halaman yang diakses dari QR code. Data dikirim via WebSocket ke backend (Gin + Gorilla WebSocket + GORM) lalu disimpan di MySQL dan dibroadcast ke dashboard admin secara realtime. Admin dapat juga mengekspor data ke Excel (.xlsx).

## Persiapan & Cara Menjalankan (Windows)
1. Salin file env:
   - Copy `.env.example` → `.env` dan isi variabel DB (DB_HOST, DB_USER, DB_PASS, DB_NAME).
2. Buat database & import skema:
   - Jalankan MySQL client:
     - mysql -u <user> -p
     - CREATE DATABASE presensi_qr;
     - exit
   - Import skema:
     - mysql -u <user> -p presensi_qr < sql/schema.sql
3. Install dependency Go:
   - buka PowerShell di folder proyek:
     - go mod tidy
4. Jalankan server:
   - go run ./cmd/server
5. Akses:
   - Form peserta: http://localhost:8080/ (atau path root yang dikonfigurasi)
   - Dashboard admin: http://localhost:8080/dashboard
   - WebSocket endpoint (backend): ws://localhost:8080/ws

Catatan: gunakan port sesuai konfigurasi di `.env` / `internal/config/config.go`.

## Contoh Pesan WebSocket (Payload)
- Dari Form (client → server):
  {
    "nama": "Budi Santoso",
    "nim": "12345678",
    "jurusan": "Teknik Informatika",
    "angkatan": 2023
  }

- Broadcast dari Server (server → dashboard clients):
  {
    "type": "update_all",
    "data": [
      { "id":1, "nama":"Budi Santoso", "nim":"12345678", "jurusan":"TI", "angkatan":2023, "waktu":"2025-11-15T08:12:00Z" },
      ...
    ],
    "total": 42
  }

Server melakukan validasi NIM duplikat sebelum INSERT; bila duplikat, server mengirim pesan error ke pengirim saja.

## Contoh Interaksi Realtime (alur singkat)
1. Peserta submit form → client JS membuka ws ke /ws dan mengirim payload.
2. Server menerima, validasi, simpan ke MySQL.
3. Setelah sukses, server mengambil daftar terbaru dan broadcast ke semua dashboard client.
4. Dashboard JS menerima pesan, render tabel baru & update counter total tanpa reload.

## Cuplikan Tampilan / Screenshot
- Tambahkan screenshot ke folder `public/screenshots/` dan sisipkan di bawah:
  [alt text](image.png)
  [alt text](image-1.png)

## Troubleshooting singkat
- Koneksi DB gagal: periksa cred di `.env`, pastikan MySQL berjalan dan skema telah diimport.
- WebSocket tidak terhubung: cek alamat ws:// dan port, pastikan server berjalan.
- Excel export: akses endpoint export (mis. /export) di dashboard atau panggil REST sesuai dokumentasi handler.

