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

package client

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"

	"code.vikunja.io/veans/internal/output"
)

type attachmentUploadResult struct {
	Success []*TaskAttachment `json:"success"`
	Errors  []struct {
		Message string `json:"message"`
	} `json:"errors"`
}

// UploadTaskAttachment uploads the file at path as an attachment to a task.
func (c *Client) UploadTaskAttachment(ctx context.Context, taskID int64, path string) (*TaskAttachment, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, output.Wrap(output.CodeValidation, err, "open %s: %v", path, err)
	}
	defer f.Close()

	var body bytes.Buffer
	mw := multipart.NewWriter(&body)
	part, err := mw.CreateFormFile("files", filepath.Base(path))
	if err != nil {
		return nil, err
	}
	if _, err := io.Copy(part, f); err != nil {
		return nil, fmt.Errorf("read %s: %w", path, err)
	}
	if err := mw.Close(); err != nil {
		return nil, err
	}

	apiPath := fmt.Sprintf("/tasks/%d/attachments", taskID)
	var res attachmentUploadResult
	if err := c.send(ctx, "POST", apiPath, nil, &body, mw.FormDataContentType(), &res); err != nil {
		return nil, err
	}
	// The server answers 201 even when the file was rejected (e.g. too large).
	if len(res.Errors) > 0 || len(res.Success) == 0 {
		msgs := make([]string, 0, len(res.Errors))
		for _, e := range res.Errors {
			msgs = append(msgs, e.Message)
		}
		return nil, output.New(output.CodeValidation, "POST %s: upload of %s rejected: %s",
			apiPath, filepath.Base(path), strings.Join(msgs, "; "))
	}
	return res.Success[0], nil
}
