FROM golang:1.21 as go-build

WORKDIR /go/src/app

COPY . ./

RUN make build-binary

FROM alpine:3.18

WORKDIR /go/src/app

COPY --from=go-build /go/src/app/bin/main ./

# Need to figure out how to expose a range of ports since FTP in passive
# mode picks one a random
EXPOSE 2023

CMD ["/go/src/app/main"]
