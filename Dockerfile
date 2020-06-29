FROM golang:1.10.4

ARG PORT
ENV GOPATH="/app"

RUN mkdir /app
ADD . /app
WORKDIR /app

ENTRYPOINT ["go", "run", "src/main.go"]

EXPOSE $PORT
