FROM golang:1.22-bookworm AS build
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=1 go build -buildvcs=false -trimpath -o /panchang-api ./cmd/panchang-api
FROM debian:bookworm-slim
RUN apt-get update && apt-get install -y --no-install-recommends ca-certificates tzdata && rm -rf /var/lib/apt/lists/*
WORKDIR /app
COPY --from=build /panchang-api ./panchang-api
COPY ephe ./ephe
ENV EPHE_PATH=/app/ephe HTTP_ADDR=:8080
EXPOSE 8080
CMD ["./panchang-api"]
