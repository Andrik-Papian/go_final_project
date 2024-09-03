FROM golang:1.22.4

WORKDIR /app

COPY . .

RUN go mod download

ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64

RUN go build -o /andrik888 ./cmd/main.go

CMD ["/andrik888"]
