# Backend (Go)
FROM golang:1.25.0-alpine AS backend-builder
WORKDIR /app
COPY govuemysql-back/ ./
RUN go build -o server .

# Frontend (Vue.js)
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY govuemysql-front/package*.json ./
RUN npm install
COPY govuemysql-front/ ./
RUN npm run build

# Final image
FROM alpine:latest
WORKDIR /app

# Copy backend binary
COPY --from=backend-builder /app/server ./server

# Copy frontend build
COPY --from=frontend-builder /app/dist ./frontend

# Expose backend port (adjust as needed)
EXPOSE 8080

CMD ["./server"]