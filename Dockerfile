FROM docker.io/golang:1.18-alpine as builder

ENV CGO_ENABLED=0

WORKDIR /src/alias-server

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY cmd cmd
COPY internal internal
COPY pkg pkg

RUN go build -o /server ./cmd/main.go

FROM scratch

COPY --from=builder /server /server

ENTRYPOINT [ "/server" ]