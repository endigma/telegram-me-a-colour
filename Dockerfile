FROM golang:1.17.1 as build-env

WORKDIR /go/src
ADD . /go/src
RUN go get -d -v ./...
RUN go build -o /go/bin/colour-srv
FROM gcr.io/distroless/base
COPY --from=build-env /go/bin/colour-srv /
COPY --from=build-env /go/src/template.html /
CMD ["/colour-srv"]
