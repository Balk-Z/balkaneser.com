FROM golang:1.21-alpine as builder

WORKDIR /app
COPY go.mod *.go ./
RUN go mod download

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-https-server


FROM scratch

COPY site ./site/
COPY --from=builder /go-https-server /go-https-server

# Run
CMD ["/go-https-server"]