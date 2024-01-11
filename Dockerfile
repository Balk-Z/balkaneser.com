FROM golang:1.21-alpine as builder

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download

# Build
COPY *.go ./
RUN CGO_ENABLED=0 GOOS=linux go build -o /go-https-server


FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go-https-server /go-https-server
COPY site ./site/


# Run
CMD ["/go-https-server"]