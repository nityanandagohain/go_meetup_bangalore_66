# Build stage
FROM golang:1.16-alpine3.13 AS builder
# enable Go modules support
ENV GO111MODULE=on

WORKDIR /app
COPY go.mod .
COPY go.sum .
RUN ls
RUN go mod download
COPY main.go .
RUN go build -o main main.go

# Run stage
FROM alpine:3.13
WORKDIR /app
COPY --from=builder /app/main .
# COPY app.env .
COPY start.sh .

EXPOSE 3000
CMD [ "/app/main" ]
ENTRYPOINT [ "/app/start.sh" ]%