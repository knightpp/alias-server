package main

import (
	"bytes"
	"encoding/json"
	"fmt"

	"google.golang.org/protobuf/encoding/protojson"
	"google.golang.org/protobuf/reflect/protoreflect"
)

func Print(v any) {
	if msg, ok := v.(protoreflect.ProtoMessage); ok {
		fmt.Println(protojson.Format(msg))
		return
	}

	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	err := enc.Encode(v)
	if err != nil {
		panic(err)
	}

	fmt.Println(buf.String())
}
