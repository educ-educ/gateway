FROM golang:1.20-alpine AS builder
WORKDIR /gateway
COPY go.mod go.sum ./
RUN go mod download && go mod verify
COPY ./ ./
WORKDIR /gateway/cmd/app
RUN CGO_ENABLED=0 go build -o /bin/app
WORKDIR /gateway

FROM alpine:latest
RUN apk --update add ca-certificates
COPY --from=builder /bin/app /bin/app
COPY --from=builder /gateway/deploy_gateway/.env /deploy_gateway/.env

CMD [ "/bin/app" ]