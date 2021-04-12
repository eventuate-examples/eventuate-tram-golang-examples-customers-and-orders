FROM golang:1.16-alpine AS build

WORKDIR /src/
COPY *.go go.* /src/
RUN CGO_ENABLED=0 go build -o /bin/customers

FROM scratch
COPY --from=build /bin/customers /bin/customers
ENTRYPOINT ["/bin/customers"]