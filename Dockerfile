FROM golang:alpine3.20 AS build

ARG MODULE
ARG MY_GITHUB_TOKEN

WORKDIR /

COPY go.mod go.sum ./
# Build with optional lambda.norpc tag
COPY ${MODULE}/main.go .
COPY ${MODULE}/service.go .

RUN apk update && apk --no-cache add git openssh-client gcc g++ mercurial && \
    git config --global url."https://${MY_GITHUB_TOKEN}:x-oauth-basic@github.com/".insteadOf "https://github.com/"

RUN go mod download
RUN CGO_ENABLED=0 GOARCH=amd64 GOOS=linux go build -tags lambda.norpc -o main *.go

# Copy artifacts to a clean image
FROM alpine:3.20 AS deploy
COPY --from=build /main ./main

ENTRYPOINT [ "./main" ]