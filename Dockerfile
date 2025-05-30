FROM golang:1.24.2-alpine AS builder
RUN apk add --no-cache make git
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN make -f Makefile.backend build
FROM alpine:latest
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app
COPY --from=builder /app/bin/backend .
COPY --from=builder /app/configs ./configs
ENV TZ=Europe/Moscow
EXPOSE 8082
CMD ["./backend"] 