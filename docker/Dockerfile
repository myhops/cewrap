# Get the latest go image
FROM docker.io/golang:latest AS builder
WORKDIR /work 
COPY go.mod go.sum ./
RUN go mod download -x
COPY . ./ 
RUN CGO_ENABLED=0 go build -o app ./cmd/source

FROM scratch
COPY --from=builder /work/app /app
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
EXPOSE 8080
ENV port=8080
ENV NAME=myipaddress_1
ENTRYPOINT ["/app"]
