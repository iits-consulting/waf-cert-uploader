FROM golang:1.21-alpine AS build

ARG CGO=0
ENV CGO_ENABLED=${CGO}
ENV GOOS=linux
ENV GO111MODULE=on

WORKDIR /go/src/github.com/iits-consulting/waf-cert-uploader
COPY . /go/src/github.com/iits-consulting/waf-cert-uploader

RUN go build -o waf-cert-uploader main.go && \
    mv waf-cert-uploader /usr/local/bin

FROM alpine:3.19
COPY --from=build /usr/local/bin/waf-cert-uploader /usr/bin/waf-cert-uploader
EXPOSE 39100
ENTRYPOINT [ "waf-cert-uploader" ]
