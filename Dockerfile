FROM golang:1.26.1-alpine AS builder

WORKDIR /build

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN go tool templ generate
RUN go build -o trady .

FROM alpine:3.23
RUN adduser -S app
USER app
EXPOSE 8080

WORKDIR /app
COPY --chown=app --from=builder /build/trady .

CMD ["./trady", "--address=0.0.0.0", "--port=8080", "--db=trady.db", "--uploads=./uploads"]
