FROM golang:1.25-alpine AS builder 

RUN apk update && apk add --no-cache git ca-certificates && update-ca-certificates

WORKDIR /go/src/golang-api-v1

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o /go/bin/server ./app/main.go

FROM scratch

COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=builder /go/bin/server /go/bin/server
ENTRYPOINT ["/go/bin/server"]
