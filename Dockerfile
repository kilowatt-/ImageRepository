FROM golang:1.14-alpine AS builder

WORKDIR /server/
COPY . .

RUN go mod download
RUN go build -o app

# build react
FROM node:alpine AS node_builder
COPY --from=builder /server/client .
RUN npm install
RUN npm run build

FROM alpine:latest
RUN apk --no-cache add ca-certificates

COPY --from=builder /server .
COPY --from=node_builder . /client
RUN chmod +x ./app

ENV PORT 25000

CMD "./app"
