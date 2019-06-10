FROM golang as builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download && go install golang.org/x/tools/cmd/stringer

COPY . /app

RUN go generate ./... && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/bot/...

FROM gcr.io/distroless/base

WORKDIR /bin/

COPY --from=builder /app/app .

WORKDIR /app/

ENTRYPOINT [ "/bin/app" ]