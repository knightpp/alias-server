FROM docker.io/golang:1.20.1-alpine as builder

ENV GIN_MODE=release

RUN go install github.com/go-task/task/v3/cmd/task@latest && \
	go clean -modcache
RUN apk add --no-cache 'binutils~=2'

WORKDIR /src/alias-server

COPY . .

RUN task build

FROM scratch

COPY --from=builder /src/alias-server/server /server

ENTRYPOINT [ "/server" ]