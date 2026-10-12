# syntax=docker/dockerfile:1

# --- Build stage: kompilasi binary statis ---
FROM golang:1.25-alpine AS build

WORKDIR /src

# Cache dependency terpisah supaya layer build tidak ter-invalidasi
# setiap kali kode aplikasi berubah.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

# CGO_ENABLED=0 menghasilkan binary statis yang jalan di scratch.
# -trimpath membuang path absolut dari binary.
RUN CGO_ENABLED=0 GOOS=linux go build \
    -trimpath -ldflags="-s -w" \
    -o /out/server ./cmd/server

# --- Runtime stage ---
FROM alpine:3.20

# ffmpeg + ffprobe wajib: dipakai untuk transcode HLS, ambil thumbnail,
# dan baca durasi video. alpine varian apk add jauh lebih kecil dari
# image Debian (~50MB vs ~250MB).
RUN apk add --no-cache ffmpeg ca-certificates tzdata && \
    adduser -D -u 10001 app

WORKDIR /app

COPY --from=build /out/server /app/server

# Folder storage DAN folder database harus dibuat di image (bukan hanya
# di runtime) supaya named volume mewarisi kepemilikan yang benar.
#
# Named volume yang ditount ke path yang tidak ada di image akan dibuat
# milik root, dan proses non-root tidak bisa menulis di dalamnya —
# gejalanya "unable to open database file: out of memory (14)".
RUN mkdir -p /var/uploads /var/hls_output /var/lib/pgarcv && \
    chown -R app:app /var/uploads /var/hls_output /var/lib/pgarcv /app

USER app

ENV ENV=production \
    PORT=8080 \
    UPLOAD_PATH=/var/uploads \
    HLS_OUTPUT_PATH=/var/hls_output \
    GIN_MODE=release

EXPOSE 8080

# Healthcheck memanggil endpoint publik supaya container ditandai unhealthy
# kalau DB atau routing bermasalah, bukan cuma kalau proses mati.
HEALTHCHECK --interval=30s --timeout=5s --start-period=20s --retries=3 \
    CMD wget -qO- http://127.0.0.1:8080/api/videos >/dev/null 2>&1 || exit 1

CMD ["/app/server"]