FROM golang:latest

WORKDIR /src
COPY . .
RUN go build -o=enguardia.bin -ldflags="-s -w"

ENTRYPOINT ["/src/enguardia.bin"]
