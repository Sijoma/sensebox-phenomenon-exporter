# Start by building the application.
FROM golang:1.16-buster as build

WORKDIR /go/src/app
ADD . /go/src/app

RUN go get -d -v ./...

RUN go build -o /go/bin/sensebox-metrics-exporter

# Now copy it into our base image.
FROM gcr.io/distroless/base-debian10
COPY --from=build /go/bin/sensebox-metrics-exporter /
CMD ["/sensebox-metrics-exporter"]