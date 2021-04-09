FROM golang:1.16
WORKDIR /go/src/app
COPY *.go ./
COPY *.mod ./
RUN go get -t .
RUN go build .
CMD ["./customers"]