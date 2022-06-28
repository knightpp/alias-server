package binding

import (
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin/binding"
	"github.com/rs/zerolog"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
)

var _ binding.Binding = (*ProtobufJSON)(nil)

type ProtobufJSON struct {
	log zerolog.Logger
}

func NewProtobuf(log zerolog.Logger) *ProtobufJSON {
	return &ProtobufJSON{}
}

func (pj *ProtobufJSON) Name() string {
	return "ProtobufJSON"
}

func (pj *ProtobufJSON) Bind(r *http.Request, out any) error {
	bodyBytes, err := io.ReadAll(r.Body)
	if err != nil {
		pj.log.Err(err).Msg("read all body")
		return fmt.Errorf("read all body: %w", err)
	}

	outMsg, ok := out.(protoreflect.ProtoMessage)
	if !ok {
		pj.log.Err(err).Msg("couldn't case to protoreflect.ProtoMessage")
		return fmt.Errorf("passed value does not implement ProtoMessage")
	}

	err = proto.Unmarshal(bodyBytes, outMsg)
	if err != nil {
		pj.log.Err(err).Msg("coudln't unmarshal proto")
		return fmt.Errorf("unmarshal proto: %w", err)
	}

	return nil
}
