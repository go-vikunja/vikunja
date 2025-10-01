package migration

import (
	"os"
	"path/filepath"
	"testing"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/user"
)

func TestMain(m *testing.M) {
	log.InitLogger()

	config.InitDefaultConfig()
	wd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	root := filepath.Dir(filepath.Dir(wd))
	if err := os.Setenv("VIKUNJA_SERVICE_ROOTPATH", root); err != nil {
		log.Fatal(err)
	}
	config.ServiceRootpath.Set(root)

	files.InitTests()
	user.InitTests()
	models.SetupTests()

	os.Exit(m.Run())
}
