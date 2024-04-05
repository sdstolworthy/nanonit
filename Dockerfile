FROM golang:1.22.1-bookworm as build

RUN apt update && apt install -y libwebp-dev libwebpmux3 libwebp7


WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

RUN go build -v -o /app/app ./...

RUN useradd -m heroku

USER heroku

CMD ["/app/app"]

