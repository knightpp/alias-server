FROM docker.io/golang:1.18-alpine as builder

ENV GIN_MODE=release

RUN go install github.com/go-task/task/v3/cmd/task@latest && \
	go clean -modcache
RUN apk add --no-cache 'binutils~=2'

WORKDIR /src/alias-server

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd cmd
COPY internal internal
COPY Taskfile.yaml .

RUN task build

FROM scratch

COPY --from=builder /src/alias-server/server /server

ENTRYPOINT [ "/server" ]