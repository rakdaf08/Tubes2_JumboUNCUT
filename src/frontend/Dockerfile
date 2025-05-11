# Tahap 1: Build aplikasi frontend
FROM node:20-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .

# Atur variabel environment untuk build

# Build aplikasi
RUN npm run build

# (Debug opsional) Cek isi folder dist
RUN ls -la /app/dist

# Tahap 2: Sajikan hasil build dengan Nginx
FROM nginx:1.27-alpine

# Copy hasil build ke direktori default Nginx
COPY --from=builder /app/dist /usr/share/nginx/html

# Copy konfigurasi nginx custom
COPY nginx.conf /etc/nginx/conf.d/default.conf

EXPOSE 3000
CMD ["nginx", "-g", "daemon off;"]
