# syntax=docker/dockerfile:1

# --- Dependencies stage ---
# Dipisah supaya `npm ci` tidak diulang setiap kali kode berubah.
FROM node:20-alpine AS deps

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

# --- Build stage ---
FROM node:20-alpine AS build

WORKDIR /app

COPY --from=deps /app/node_modules ./node_modules
COPY . .

# NEXT_PUBLIC_* di-inline ke bundle browser saat build. Dipakai string kosong
# supaya browser memakai path relatif ("/api/...", "/uploads/...") yang
# ditangani rewrite Next.js — dengan begitu frontend tidak perlu tahu
# alamat internal container backend.
#
# API_URL hanya untuk sisi server (rewrite + proxy). Harus berupa build arg
# karena rewrites dievaluasi saat build, bukan saat runtime.
ARG API_URL=http://backend:8080

ENV NEXT_TELEMETRY_DISABLED=1 \
    NEXT_PUBLIC_API_URL="" \
    API_URL=${API_URL}

RUN npm run build

# --- Runtime stage ---
FROM node:20-alpine AS runner

WORKDIR /app

ENV NODE_ENV=production \
    NEXT_TELEMETRY_DISABLED=1 \
    PORT=3000

# next.config.js dipakai runtime untuk evaluation rewrite, jadi harus ikut.
COPY --from=build /app/next.config.js ./next.config.js
COPY --from=build /app/public ./public
COPY --from=build /app/.next ./.next
COPY --from=build /app/node_modules ./node_modules
COPY --from=build /app/package.json ./package.json

RUN addgroup -g 10001 -S nodejs && \
    adduser -S nextjs -u 10001 && \
    chown -R nextjs:nodejs /app/.next

USER nextjs

EXPOSE 3000

HEALTHCHECK --interval=30s --timeout=5s --start-period=30s --retries=3 \
    CMD wget -qO- http://127.0.0.1:3000/ >/dev/null 2>&1 || exit 1

CMD ["npm", "run", "start"]