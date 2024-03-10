FROM golang:1.22.1-bookworm

RUN apt update && apt install -y libwebp-dev libwebpmux3 libwebp7


WORKDIR /app

COPY go.mod go.sum ./

RUN go mod download && go mod verify

COPY . .

RUN go build -v -o /app/app ./...

RUN git clone --depth 1 -b main https://github.com/tidbyt/community.git /community

ARG GOOGLE_APPLICATION_CREDENTIALS

ENV GOOGLE_APPLICATION_CREDENTIALS="/app/credentials.json"
EXPOSE 8080
ENV APPS_PATH="/community/apps"
CMD ["/app/app"]

