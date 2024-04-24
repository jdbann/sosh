FROM golang:1.21.6

WORKDIR /usr/src/app

COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app .

EXPOSE 23234/tcp

CMD ["app"]
