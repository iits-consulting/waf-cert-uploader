FROM golang:1.21.5

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /waf-cert-uploader ./

CMD ["/waf-cert-uploader"]