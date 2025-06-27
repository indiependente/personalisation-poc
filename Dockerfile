# Build stage
FROM golang:latest AS base
WORKDIR /app
ADD go.mod .
ADD go.sum .
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o main .

# Production stage
FROM gcr.io/distroless/static
COPY --from=base /app/main /app/main
EXPOSE 8080
ENTRYPOINT [ "/app/main" ]
