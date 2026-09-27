FROM golang:1.27.1 AS builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w" -o /out/lifnex ./cmd/lifnex

FROM gcr.io/distroless/static-debian13:nonroot

COPY --from=builder /out/lifnex /lifnex
USER nonroot:nonroot
ENTRYPOINT ["/lifnex"]
