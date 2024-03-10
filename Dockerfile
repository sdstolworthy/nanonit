FROM golang:1.22.1-bookworm

RUN apt update && apt install -y libwebp-dev libwebpmux3 libwebp7


WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .


ENV GOOGLE_APPLICATION_CREDENTIALS=""
EXPOSE 8080
ENV APPS_PATH="./tidbytcommunity/apps"

RUN go build -v -o /app/app ./...

CMD ["/app/app"]

