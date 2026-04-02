FROM golang:1.25.3-alpine AS builder

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/fanone-video .

FROM alpine:3.22

WORKDIR /app

RUN apk add --no-cache ca-certificates tzdata

COPY --from=builder /out/fanone-video /app/fanone-video

RUN mkdir -p /app/storage/avatars /app/storage/videos

EXPOSE 8888

ENTRYPOINT ["/app/fanone-video"]
