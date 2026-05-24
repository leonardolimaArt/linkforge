FROM golang:1.26-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w" \
    -o /out/redirect \
    ./cmd/redirect


FROM gcr.io/distroless/static:nonroot
WORKDIR /app
COPY --from=builder /out/redirect /app/redirect
USER nonroot:nonroot
EXPOSE 8081
ENTRYPOINT ["/app/redirect"]