# ---------- Build stage ----------
FROM golang:1.22-alpine

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod tidy

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o app ./main.go

# ---------- Runtime stage ----------
FROM alpine:latest

WORKDIR /app

COPY --from=builder /app/app .

EXPOSE 8080

CMD ["./app"]