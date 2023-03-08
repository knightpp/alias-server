package socket_test

import (
	"testing"

	"github.com/knightpp/alias-server/internal/testutil/testserver"
	"github.com/knightpp/alias-server/internal/uuidgen"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
)

func TestSocket(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Socket Suite")
}

var _ = BeforeSuite(func() {
	uuidgen.SetGlobal(uuidgen.NewConstant(testserver.TestUUID))
})
