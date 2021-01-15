FROM golang:1.14-alpine AS builder

WORKDIR /server/
COPY . .

RUN go mod download
RUN go build -o app

FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=builder /server .
RUN chmod +x ./app

ENV GO_PORT 25000

CMD "./app"
