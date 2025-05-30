# Stage 1: build binary
FROM golang:1.23.9 AS builder

WORKDIR /room
COPY . .
RUN go build -o bedroompop

# Stage 2: buat image ringan
FROM debian:bookworm-slim

WORKDIR /room
COPY --from=builder /room/bedroompop .

EXPOSE 7000
EXPOSE 7070
CMD ["./bedroompop"]
