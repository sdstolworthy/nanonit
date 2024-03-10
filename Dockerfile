FROM golang:1.22.1-bookworm

RUN apt update && apt install -y libwebp-dev libwebpmux3 libwebp7

WORKDIR /app

COPY go.mod go.sum ./
RUN ls && go mod download && go mod verify

COPY . .
RUN go build -v -o /app/app ./...

CMD ["app"]

