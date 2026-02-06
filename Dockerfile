FROM golang:1.25.7-alpine AS build
WORKDIR /app
COPY go.mod .
RUN go mod download
COPY . .
RUN ./scripts/sqlc_generate.sh
RUN CGO_ENABLED=0 GOOS=linux go build -o /bin/api ./cmd/api

FROM gcr.io/distroless/static-debian12
WORKDIR /app
COPY --from=build /bin/api /api
COPY db/migrations ./db/migrations
COPY api/openapi.yaml ./api/openapi.yaml
ENV HTTP_ADDR=:8080
EXPOSE 8080
ENTRYPOINT ["/api"]
