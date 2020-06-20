// Vikunja is a to-do list application to facilitate your life.
// Copyright 2018-2020 Vikunja and contributors. All rights reserved.
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package dump

import (
	"archive/zip"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/version"
	"fmt"
	"github.com/spf13/viper"
	"io"
	"os"
	"strconv"
)

// Change to deflate to gain better compression
// see http://golang.org/pkg/archive/zip/#pkg-constants
const compressionUsed = zip.Deflate

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
	err = writeBytesToZip("VERSION", []byte(version.Version), dumpWriter)
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
		err = writeBytesToZip("database/"+t+".json", d, dumpWriter)
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
	for fid, fcontent := range allFiles {
		err = writeBytesToZip("files/"+strconv.FormatInt(fid, 10), fcontent, dumpWriter)
		if err != nil {
			return fmt.Errorf("error writing file %d: %s", fid, err)
		}
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
	header.Method = compressionUsed

	w, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(w, fileToZip)
	return err
}

func writeBytesToZip(filename string, data []byte, writer *zip.Writer) (err error) {
	header := &zip.FileHeader{
		Name:   filename,
		Method: compressionUsed,
	}
	w, err := writer.CreateHeader(header)
	if err != nil {
		return err
	}
	_, err = w.Write(data)
	return
}
