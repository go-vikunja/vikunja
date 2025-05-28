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

package notifications

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/version"
)

// WebhookPayload represents the payload sent to webhook URLs
type WebhookPayload struct {
	EventName string      `json:"event_name"`
	Time      time.Time   `json:"time"`
	Data      interface{} `json:"data"`
}

var webhookClient *http.Client

func getWebhookHTTPClient() *http.Client {
	if webhookClient != nil {
		return webhookClient
	}

	client := &http.Client{}
	client.Timeout = time.Duration(config.WebhooksTimeoutSeconds.GetInt()) * time.Second

	if config.WebhooksProxyURL.GetString() == "" || config.WebhooksProxyPassword.GetString() == "" {
		webhookClient = client
		return webhookClient
	}

	proxyURL, _ := url.Parse(config.WebhooksProxyURL.GetString())

	client.Transport = &http.Transport{
		Proxy: http.ProxyURL(proxyURL),
		ProxyConnectHeader: http.Header{
			"Proxy-Authorization": []string{
				"Basic " + base64.StdEncoding.EncodeToString(
					[]byte(config.WebhooksProxyURL.GetString()+":"+config.WebhooksProxyPassword.GetString()),
				),
			},
		},
	}

	webhookClient = client
	return webhookClient
}

// sendWebhookPayload sends a webhook payload to the specified URL
func sendWebhookPayload(targetURL string, p *WebhookPayload) error {
	payload, err := json.Marshal(p)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodPost, targetURL, bytes.NewReader(payload))
	if err != nil {
		return err
	}

	req.Header.Add("User-Agent", "Vikunja/"+version.Version)
	req.Header.Add("Content-Type", "application/json")

	client := getWebhookHTTPClient()
	res, err := client.Do(req)
	if err != nil {
		return err
	}

	defer res.Body.Close()

	if res.StatusCode > 399 {
		responseBody, err := io.ReadAll(res.Body)
		if err != nil {
			return err
		}

		log.Errorf("Got response with status %d from webhook URL %s: %s", res.StatusCode, targetURL, responseBody)
	}

	log.Debugf("Sent webhook payload to %s for event %s", targetURL, p.EventName)
	return nil
}
