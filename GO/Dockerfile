# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o barcvvr .

# Run stage
FROM alpine:latest
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /app/barcvvr .
COPY --from=builder /app/templates ./templates
COPY --from=builder /app/assets ./assets

# To satisfy Gin's LoadHTMLGlob("templates/*/*.html") limiting regex:
RUN cp templates/layout.html templates/users/ && \
    cp templates/layout.html templates/drinks/ && \
    cp templates/layout.html templates/kegs/ && \
    cp templates/layout.html templates/operations/ && \
    cp templates/layout.html templates/beerflows/

EXPOSE 3000
ENV GIN_MODE=release
ENV PORT=3000

CMD ["./barcvvr"]
