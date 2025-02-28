FROM golang:1.23.3-bookworm


COPY . /go/src
WORKDIR /go/src/_examples/http-server
RUN cp /go/src/filter/traceable/libs/linux_arm64/libtraceable.so /go/src/

RUN go mod tidy
RUN go build -tags=traceable_filter -o /go/src/app.o .

ENTRYPOINT ["/go/src/app.o"]