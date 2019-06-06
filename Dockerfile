FROM golang as builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download

COPY . /app

RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/bot/...

# deployment image
FROM scratch

# copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

WORKDIR /bin/

COPY --from=builder /app/app .

WORKDIR /app/

ENTRYPOINT [ "/bin/app" ]