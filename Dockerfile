FROM docker.io/golang:1.20.3-alpine as builder

RUN go install github.com/go-task/task/v3/cmd/task@latest && \
	go clean -modcache
RUN apk add --no-cache 'binutils~=2'

WORKDIR /src/alias-server

COPY . .

ENV CGO_ENABLED="0"
RUN task build

FROM docker.io/alpine:3.17

COPY --from=builder /src/alias-server/server /server

ENTRYPOINT [ "/server" ]
