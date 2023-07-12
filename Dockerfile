FROM golang:1.20 as builder

WORKDIR /devnull

COPY go.mod ./

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /devnull

RUN chmod +x /devnull

EXPOSE 8080

CMD ["/devnull/devnull"]