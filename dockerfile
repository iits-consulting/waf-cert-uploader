FROM golang:1.21

WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /waf-webhook ./

CMD ["/waf-webhook"]