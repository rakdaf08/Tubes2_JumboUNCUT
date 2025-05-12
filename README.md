<h1 align="center">Tugas Besar 2 IF2211 Strategi Algoritma</h1>
<h1 align="center">Pencarian Resep dalam Permainan Alchemy menggunakan Algoritma BFS, DFS, dan BDS</h1>

![Jumbo Uncut Logo](src/frontend/public/jumbo-kocok-logo.svg)

## Daftar Isi

1. [Informasi Umum](#informasi-umum)
2. [Kontributor](#kontributor)
3. [Fitur](#fitur)
4. [Requirements Program](#requirements-program)
5. [Cara Menjalankan Program](#cara-menjalankan-program)
6. [Status Proyek](#status-proyek)
7. [Struktur Proyek](#struktur-proyek)

## Informasi Umum

Aplikasi ini merupakan implementasi dari algoritma pencarian seperti Breadth-First Search (BFS), Depth-First Search (DFS), dan Bidirectional Search (BDS) untuk mencari resep pembuatan elemen dalam permainan Alchemy. Pengguna dapat memilih elemen target, algoritma pencarian, dan mode pencarian (jalur terpendek atau beberapa jalur) untuk menemukan cara terbaik membuat elemen dari elemen dasar.

## Kontributor

### **Kelompok JumboUNCUT**

|   NIM    |           Nama            |
| :------: | :-----------------------: |
| 13523018 |    Raka Daffa Iftihaar    |
| 13523038 | Abrar Abhirama Widyadhana |
| 13523055 |  Muhammad Timur Kanigara  |

## Fitur

Fitur yang digunakan dalam program ini:
| NO | Algoritma | Deskripsi |
|:---:|----------------------|----------------------------------------------------------------------|
| 1 | BFS | Pencarian jalur resep menggunakan algoritma Breadth First Search |
| 2 | DFS | Pencarian jalur resep menggunakan algoritma Depth First Search |
| 3 | BDS | Pencarian jalur resep menggunakan algoritma Bidirectional Search |

## Requirements Program

| NO  | Required Program | Reference Link                                            |
| :-: | ---------------- | --------------------------------------------------------- |
|  1  | Node.js dan npm  | [Node.js](https://nodejs.org/)                            |
|  2  | Go Language      | [The Go Programming Language](https://go.dev)             |
|  3  | React + Vite     | [React](https://react.dev) + [Vite](https://vitejs.dev/)  |
|  4  | Docker Desktop   | [Docker](https://www.docker.com/products/docker-desktop/) |

## Cara Menjalankan Program

### Menggunakan Docker (Rekomendasi)

1. Clone repository ini dengan perintah `git clone https://github.com/rakdaf08/Tubes2_JumboUNCUT.git`
2. Pastikan Docker Desktop sudah berjalan di komputer Anda
3. Masuk ke direktori `src` dengan perintah `cd src`
4. Jalankan perintah `docker-compose up --build` untuk membangun dan menjalankan aplikasi
5. Buka `http://localhost:3000` di browser Anda
6. Untuk menghentikan aplikasi, jalankan perintah `docker-compose down`

### Tanpa Docker (Development)

#### Backend

1. Pastikan Go sudah terinstall
2. Masuk ke direktori backend: `cd src/backend`
3. Download semua dependensi: `go mod download`
4. Jalankan backend: `go run .`
5. Backend akan berjalan di `http://localhost:8080`

#### Frontend

1. Pastikan Node.js dan npm sudah terinstall
2. Masuk ke direktori frontend: `cd src/frontend`
3. Install semua dependensi: `npm install`
4. Jalankan frontend: `npm run dev`
5. Frontend akan berjalan di `http://localhost:3000`

### Cara Menggunakan Website

1. Buka halaman utama aplikasi di `http://localhost:3000` atau jika ingin menggunakan link deployment bisa menuju ke `https://tubes2-jumbo-uncut.vercel.app`
2. Pilih elemen target yang ingin dibuat
3. Pilih algoritma pencarian (BFS, DFS, atau BDS)
4. Pilih mode pencarian (Single untuk jalur terpendek, Multiple untuk beberapa jalur)
5. Jika memilih mode Multiple, tentukan jumlah jalur maksimum yang ingin ditampilkan
6. Klik tombol "Cari Resep" dan tunggu hasil pencarian muncul
7. Hasil akan ditampilkan dalam bentuk visualisasi graph dan daftar langkah-langkah

## Status Proyek

Proyek ini telah selesai dan siap digunakan.

## Struktur Proyek

```bash
Tubes2_JumboUNCUT/
├── [README.md](http://_vscodecontentref_/0)
├── src/
│   ├── backend/
│   │   ├── bfs.go          # Implementasi algoritma BFS
│   │   ├── dfs.go          # Implementasi algoritma DFS
│   │   ├── bds.go          # Implementasi algoritma BDS
│   │   ├── data/           # Data elemen dan resep
│   │   ├── handlers.go     # Handler API
│   │   ├── main.go         # Entry point backend
|   |   ├── Dockerfile      # Docker untuk backend
│   │   └── ...
│   ├── frontend/
│   │   ├── public/         # Aset publik
│   │   ├── src/
│   │   │   ├── api/        # Layanan API
│   │   │   ├── components/ # Komponen React
│   │   │   ├── pages/      # Halaman utama
│   │   │   ├── App.jsx     # Komponen utama
│   │   │   └── ...
│   │   ├── [package.json](http://_vscodecontentref_/1)
│   │   ├── vite.config.js
|   |   ├── Dockerfile      # Docker untuk frontend
│   │   └── ...
│   ├── docker-compose.yml  # Konfigurasi Docker Compose
│   └── ...
└── docs/                   # Dokumentasi
    └── ...
```
