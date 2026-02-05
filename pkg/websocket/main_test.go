package websocket

import (
	"os"
	"testing"

	"code.vikunja.io/api/pkg/log"
)

func TestMain(m *testing.M) {
	log.InitLogger()
	os.Exit(m.Run())
}
