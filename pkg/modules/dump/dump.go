// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-present Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public License for more details.
//
// You should have received a copy of the GNU Affero General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package dump

import (
	"archive/zip"
	"fmt"
	"io"
	"os"
	"strings"

	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/utils"
	"code.vikunja.io/api/pkg/version"
	"github.com/spf13/viper"
)

// Dump creates a zip file with all vikunja files at filename
func Dump(filename string) error {
	dumpFile, err := os.Create(filename)
	if err != nil {
		return fmt.Errorf("error opening dump file: %w", err)
	}
	defer dumpFile.Close()

	dumpWriter := zip.NewWriter(dumpFile)
	defer dumpWriter.Close()

	// Config
	log.Info("Start dumping config file...")
	if viper.ConfigFileUsed() != "" {
		err = writeFileToZip(viper.ConfigFileUsed(), dumpWriter)
		if err != nil {
			return fmt.Errorf("error saving config file: %w", err)
		}
	} else {
		log.Warning("No config file found, not including one in the dump. This usually happens when environment variables are used for configuration.")
	}
	log.Info("Dumped config file")

	env := os.Environ()
	dotEnv := ""
	for _, e := range env {
		if strings.Contains(e, "VIKUNJA_") {
			dotEnv += e + "\n"
		}
	}
	if dotEnv != "" {
		err = utils.WriteBytesToZip(".env", []byte(dotEnv), dumpWriter)
		if err != nil {
			return fmt.Errorf("error saving env file: %w", err)
		}
		log.Info("Dumped .env file")
	}

	// Version
	log.Info("Start dumping version file...")
	err = utils.WriteBytesToZip("VERSION", []byte(version.Version), dumpWriter)
	if err != nil {
		return fmt.Errorf("error saving version: %w", err)
	}
	log.Info("Dumped version")

	// Database
	log.Info("Start dumping database...")
	data, err := db.Dump()
	if err != nil {
		return fmt.Errorf("error saving database data: %w", err)
	}
	for t, d := range data {
		err = utils.WriteBytesToZip("database/"+t+".json", d, dumpWriter)
		if err != nil {
			return fmt.Errorf("error writing database table %s: %w", t, err)
		}
	}
	log.Info("Dumped database")

	// Files
	log.Info("Start dumping files...")
	allFiles, err := files.Dump()
	if err != nil {
		return fmt.Errorf("error saving file: %w", err)
	}

	err = utils.WriteFilesToZip(allFiles, dumpWriter)
	if err != nil {
		return err
	}

	log.Infof("Dumped files")

	log.Info("Done creating dump")
	log.Infof("Dump file saved at %s", filename)
	return nil
}

func writeFileToZip(filename string, writer *zip.Writer) error {
	// #nosec
	fileToZip, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer fileToZip.Close()

	// Get the file information
	info, err := fileToZip.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	header.Name = info.Name()
	header.Method = utils.CompressionUsed

	w, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, fileToZip)
	return err
}
