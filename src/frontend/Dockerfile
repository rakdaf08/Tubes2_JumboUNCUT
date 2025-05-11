# Tahap 1: Build aplikasi frontend (menggunakan Node.js)
# Gunakan versi Node.js yang sesuai dengan proyek Anda.
# Node.js 20 LTS dengan base Alpine direkomendasikan untuk ukuran dan keamanan.
FROM node:20-alpine AS builder

# Tetapkan direktori kerja di dalam container builder
WORKDIR /app

# Salin package.json dan package-lock.json (atau yarn.lock jika menggunakan Yarn)
# Menyalin file-file ini terlebih dahulu memungkinkan Docker untuk men-cache lapisan dependensi
# jika file-file ini tidak berubah, sehingga mempercepat build berikutnya.
COPY package.json package-lock.json ./
# Jika menggunakan Yarn:
# COPY package.json yarn.lock ./

# Instal dependensi proyek
# 'npm ci' direkomendasikan untuk build yang lebih cepat dan reproduktif berdasarkan package-lock.json
RUN npm ci
# Jika menggunakan Yarn:
# RUN yarn install --frozen-lockfile

# Salin sisa file aplikasi frontend ke direktori kerja di container
# Ini termasuk direktori src, public, vite.config.js, dll.
COPY . .

# Build aplikasi untuk produksi
# Perintah ini biasanya 'npm run build' untuk proyek Vite (sesuaikan jika berbeda di package.json Anda)
RUN npm run build
# Ini akan menghasilkan file statis di direktori '/app/dist' (output default Vite)

# Tahap 2: Sajikan aplikasi menggunakan server web ringan (Nginx)
# Menggunakan image Nginx berbasis Alpine yang sangat kecil
FROM nginx:1.27-alpine
# Anda bisa menggunakan versi Nginx yang lebih baru jika tersedia dan sesuai

# Salin file build statis dari tahap 'builder' ke direktori default Nginx untuk disajikan
# Pastikan path sumber '/app/dist' sesuai dengan output build Vite Anda.
# Direktori target '/usr/share/nginx/html' adalah root dokumen default Nginx.
COPY --from=builder /app/dist /usr/share/nginx/html

# (Opsional) Jika Anda memiliki file konfigurasi Nginx kustom (misalnya untuk routing SPA):
# Buat file bernama 'nginx.conf' atau 'default.conf' di direktori frontend Anda, lalu salin:
# COPY nginx-custom.conf /etc/nginx/conf.d/default.conf

# Ekspos port yang digunakan oleh Nginx (default adalah 80)
EXPOSE 80

# Perintah default Nginx untuk menjalankan server di foreground
CMD ["nginx", "-g", "daemon off;"]