FROM golang:1.20 as build

WORKDIR /go/src/app

COPY go.mod ./
COPY *.go ./

RUN CGO_ENABLED=0 go build -o /go/bin/app

FROM gcr.io/distroless/static-debian12

COPY --from=build /go/bin/app /
COPY assets/ /assets/

EXPOSE 8080

CMD ["/app", "serve"]