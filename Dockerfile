FROM golang as builder

WORKDIR /app

COPY go.mod go.sum /app/

RUN go mod download && go install golang.org/x/tools/cmd/stringer

COPY . /app

RUN go generate ./... && CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o app ./cmd/bot/...

FROM golang:latest

WORKDIR /bin/

COPY --from=builder /app/app .

COPY tmpl /tmpl/

WORKDIR /app/

ENTRYPOINT [ "/bin/app" ]