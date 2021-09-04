// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2021 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU Affero General Public Licensee as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU Affero General Public Licensee for more details.
//
// You should have received a copy of the GNU Affero General Public Licensee
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package dump

import (
	"archive/zip"
	"fmt"
	"io"
	"os"

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
		return fmt.Errorf("error opening dump file: %s", err)
	}
	defer dumpFile.Close()

	dumpWriter := zip.NewWriter(dumpFile)
	defer dumpWriter.Close()

	// Config
	log.Info("Start dumping config file...")
	err = writeFileToZip(viper.ConfigFileUsed(), dumpWriter)
	if err != nil {
		return fmt.Errorf("error saving config file: %s", err)
	}
	log.Info("Dumped config file")

	// Version
	log.Info("Start dumping version file...")
	err = utils.WriteBytesToZip("VERSION", []byte(version.Version), dumpWriter)
	if err != nil {
		return fmt.Errorf("error saving version: %s", err)
	}
	log.Info("Dumped version")

	// Database
	log.Info("Start dumping database...")
	data, err := db.Dump()
	if err != nil {
		return fmt.Errorf("error saving database data: %s", err)
	}
	for t, d := range data {
		err = utils.WriteBytesToZip("database/"+t+".json", d, dumpWriter)
		if err != nil {
			return fmt.Errorf("error writing database table %s: %s", t, err)
		}
	}
	log.Info("Dumped database")

	// Files
	log.Info("Start dumping files...")
	allFiles, err := files.Dump()
	if err != nil {
		return fmt.Errorf("error saving file: %s", err)
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
