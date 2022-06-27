package main

import (
	"bytes"
	"io"
	"net/http"

	serverpb "github.com/knightpp/alias-proto/go/pkg/server/v1"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

const baseURL = "http://127.0.0.1:8080"

func main() {
	resp := createRoom(&serverpb.CreateRoomRequest{
		Name:     "My test room",
		IsPublic: false,
		Language: "UKR",
		Password: nil,
	})

	Print(resp)
}

func createRoom(req *serverpb.CreateRoomRequest) *serverpb.CreateRoomResponse {
	body := bytes.NewReader(b(req))

	httpReq, err := http.NewRequest("POST", baseURL+"/room", body)
	if err != nil {
		panic(err)
	}

	httpReq.Header.Set("auth", "3e7a110c-1832-4e73-979a-8451022709a6")

	// httpReq.
	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		panic(err)
	}

	var room serverpb.CreateRoomResponse
	msgFromReader(resp.Body, &room)

	return &room
}

func msgFromReader(r io.Reader, m proto.Message) {
	bytes, err := io.ReadAll(r)
	if err != nil {
		panic(err)
	}

	err = proto.Unmarshal(bytes, m)
	if err != nil {
		panic(err)
	}
}

func b(m protoreflect.ProtoMessage) []byte {
	bytes, err := proto.Marshal(m)
	if err != nil {
		panic(err.Error())
	}
	return bytes
}
