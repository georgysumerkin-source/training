FROM golang:1.26-alpine AS builder

WORKDIR /app

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o /server .

FROM scratch

COPY --from=builder /server /server

EXPOSE 9090

ENTRYPOINT ["/server"]