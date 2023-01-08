FROM golang:latest

COPY . /app

WORKDIR /app

RUN go get -d -v ./...
RUN go install -v ./...

EXPOSE 8080

CMD ["app"]
