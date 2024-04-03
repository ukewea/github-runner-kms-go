FROM golang:1.22 as builder

ARG GOARCH=amd64

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=${GOARCH} go build -o main .

FROM alpine:latest
LABEL org.opencontainers.image.source="https://github.com/ukewea/github-runner-kms-go"

WORKDIR /app/
COPY --from=builder /app/main .
EXPOSE 3000

CMD ["./main"]