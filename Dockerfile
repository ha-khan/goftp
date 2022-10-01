FROM golang:1.19 as go-build

WORKDIR /go/src/app

COPY . ./

RUN make build-binary

FROM alpine:3.15

WORKDIR /go/src/app

COPY --from=go-build /go/src/app/bin/main ./

EXPOSE 2023

CMD ["/go/src/app/main"]
