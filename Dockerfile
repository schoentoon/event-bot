FROM golang:alpine as builder

RUN apk --no-cache add tzdata zip ca-certificates git

WORKDIR /usr/share/zoneinfo

# -0 means no compression.  Needed because go's
# tz loader doesn't handle compressed data.
RUN zip -r -0 /zoneinfo.zip .

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download && go install golang.org/x/tools/cmd/stringer

COPY . /app

RUN go generate ./... && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/bot/...

# deployment image
FROM scratch

# copy ca-certificates from builder
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt

# copy zoneinfo from builder
COPY --from=builder /zoneinfo.zip /

WORKDIR /bin/

COPY --from=builder /app/app .

WORKDIR /app/

ENTRYPOINT [ "/bin/app" ]