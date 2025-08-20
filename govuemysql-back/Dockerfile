# build
FROM golang:1.25.0 AS build
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
# run tests inside the image build
RUN go test ./...
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/server ./cmd/server

# run
FROM gcr.io/distroless/base-debian12
WORKDIR /app
COPY --from=build /app/server /app/server
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/app/server"]
