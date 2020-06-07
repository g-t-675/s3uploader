FROM golang:1.14.3-buster as build

LABEL maintainer=goce.trenchev@gmail.com

RUN mkdir /app
ADD . /app/
WORKDIR /app
RUN set -x && \
    go get github.com/aws/aws-sdk-go/...

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o /app/mainapp

FROM gcr.io/distroless/base-debian10
COPY --from=build /app/mainapp /mainapp
ENTRYPOINT ["/mainapp"]
