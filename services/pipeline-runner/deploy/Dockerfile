FROM golang:alpine AS build

WORKDIR /go/src/app

COPY . .

RUN go build -o build/app ./...

FROM alpine AS run

WORKDIR /opt

COPY --from=build /go/src/app/build/app .

EXPOSE 8080

ENTRYPOINT ["./app"]
