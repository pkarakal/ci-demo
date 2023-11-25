FROM golang:1.21.3-alpine as build

WORKDIR /app
COPY go.mod go.sum /app/
RUN go mod download

COPY . .
RUN go build

FROM alpine:3.18.4 as final

USER 1000
COPY --from=build /app/demo /demo

ENTRYPOINT ["/demo"]
