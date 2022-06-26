FROM docker.io/golang:1.18-alpine as builder

ENV CGO_ENABLED=0
ENV GIN_MODE=release

WORKDIR /src/alias-server

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd cmd
COPY internal internal

RUN go build -o /server -tags=nomsgpac ./cmd/main.go

FROM scratch

COPY --from=builder /server /server

ENTRYPOINT [ "/server" ]