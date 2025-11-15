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

package deck

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"code.vikunja.io/api/pkg/config"
	"code.vikunja.io/api/pkg/db"
	"code.vikunja.io/api/pkg/files"
	"code.vikunja.io/api/pkg/log"
	"code.vikunja.io/api/pkg/models"
	"code.vikunja.io/api/pkg/modules/migration"
	"code.vikunja.io/api/pkg/user"
	"github.com/yuin/goldmark"
	"xorm.io/xorm"
)

const apiVersion = "v1.0"

// Migration is the Nextcloud Deck migration struct
type Migration struct {
	Code      string `json:"code" valid:"required" minLength:"1" maxLength:"500"`
	ServerURL string `json:"server_url" valid:"required" minLength:"10" maxLength:"250"`

	// userMappingCache caches the loaded user mapping from config
	// Maps Nextcloud username -> Vikunja user
	userMappingCache map[string]*user.User
}

// Deck API response structs

type deckBoard struct {
	ID           int         `json:"id"`
	Title        string      `json:"title"`
	Color        string      `json:"color"`
	Archived     bool        `json:"archived"`
	Labels       []deckLabel `json:"labels"`
	DeletedAt    int64       `json:"deletedAt"`
	LastModified int64       `json:"lastModified"`
}

type deckStack struct {
	ID           int        `json:"id"`
	Title        string     `json:"title"`
	BoardID      int        `json:"boardId"`
	Order        int        `json:"order"`
	Cards        []deckCard `json:"cards"`
	DeletedAt    int64      `json:"deletedAt"`
	LastModified int64      `json:"lastModified"`
}

type deckCard struct {
	ID            int                `json:"id"`
	Title         string             `json:"title"`
	Description   string             `json:"description"`
	StackID       int                `json:"stackId"`
	Order         int                `json:"order"`
	Archived      bool               `json:"archived"`
	DueDate       string             `json:"duedate"`
	DeletedAt     int64              `json:"deletedAt"`
	CreatedAt     int64              `json:"createdAt"`
	LastModified  int64              `json:"lastModified"`
	Labels        []deckLabel        `json:"labels"`
	AssignedUsers []deckAssignedUser `json:"assignedUsers"`
	Done          *string            `json:"done"` // ISO-8601 datetime string when done, null otherwise
}

type deckLabel struct {
	ID        int    `json:"id"`
	Title     string `json:"title"`
	Color     string `json:"color"`
	BoardID   int    `json:"boardId"`
	DeletedAt int64  `json:"deletedAt"`
}

type deckComment struct {
	ID               int    `json:"id"`
	ObjectID         int    `json:"objectId"`
	ActorID          string `json:"actorId"`
	ActorDisplayName string `json:"actorDisplayName"`
	CreationDateTime string `json:"creationDateTime"`
	Message          string `json:"message"`
}

// OCS API response wrapper types
type ocsResponse struct {
	OCS *ocsData `json:"ocs"`
}

type ocsData struct {
	Meta *ocsMeta        `json:"meta"`
	Data json.RawMessage `json:"data"`
}

type ocsMeta struct {
	Status     string `json:"status"`
	StatusCode int    `json:"statuscode"`
	Message    string `json:"message"`
}

type deckAttachment struct {
	ID           int                   `json:"id"`
	CardID       int                   `json:"cardId"`
	Data         string                `json:"data"`
	Type         string                `json:"type"`
	CreatedAt    int64                 `json:"createdAt"`
	CreatedBy    string                `json:"createdBy"`
	DeletedAt    int64                 `json:"deletedAt"`
	ExtendedData deckAttachmentExtData `json:"extendedData"`
}

type deckAttachmentExtData struct {
	Filesize          int64                 `json:"filesize"`
	Mimetype          string                `json:"mimetype"`
	FileID            int64                 `json:"fileid"`            // Nextcloud file ID for direct download
	Path              string                `json:"path"`              // File path in Nextcloud
	Data              string                `json:"data"`              // Filename
	AttachmentCreator deckAttachmentCreator `json:"attachmentCreator"` // File creator info with username
}

type deckAttachmentCreator struct {
	ID          string `json:"id"`          // Username/UID
	DisplayName string `json:"displayName"` // Display name
	Email       string `json:"email"`       // Email
}

type deckAssignedUser struct {
	Participant deckUser `json:"participant"`
}

type deckUser struct {
	UID         string `json:"uid"`
	DisplayName string `json:"displayname"`
}

// oauthTokenResponse represents the response from Nextcloud OAuth token endpoint
type oauthTokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

// Name returns the migrator name
// @Summary Get migration status
// @Description Returns if the current user already did the migation or not
// @tags migration
// @Produce json
// @Security JWTKeyAuth
// @Success 200 {object} migration.Status "The migration status"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/deck/status [get]
func (m *Migration) Name() string {
	return "deck"
}

// AuthURL generates and returns the OAuth authorization URL for Nextcloud Deck
// @Summary Get OAuth authorization URL for Nextcloud Deck
// @Description Validates the server URL and returns the OAuth authorization URL
// @tags migration
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param serverURL body string true "Nextcloud server URL"
// @Success 200 {object} handler.AuthURL "The OAuth authorization URL"
// @Failure 400 {object} models.Message "Bad request"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/deck/auth [post]
func (m *Migration) AuthURL() string {
	// Validate server URL
	if err := validateNextcloudServer(m.ServerURL); err != nil {
		log.Errorf("[Deck Migration] Server validation failed: %v", err)
		return ""
	}

	// Get OAuth configuration
	clientID := config.MigrationDeckClientID.GetString()
	redirectURL := config.MigrationDeckRedirectURL.GetString()

	if clientID == "" || redirectURL == "" {
		log.Error("[Deck Migration] OAuth not configured: missing client ID or redirect URL")
		return ""
	}

	// Generate OAuth state for CSRF protection
	state, err := generateOAuthState()
	if err != nil {
		log.Errorf("[Deck Migration] Failed to generate OAuth state: %v", err)
		return ""
	}

	// Build authorization URL
	serverURL := strings.TrimRight(m.ServerURL, "/")
	authURL := fmt.Sprintf(
		"%s/index.php/apps/oauth2/authorize?client_id=%s&redirect_uri=%s&response_type=code&state=%s",
		serverURL,
		url.QueryEscape(clientID),
		url.QueryEscape(redirectURL),
		url.QueryEscape(state),
	)

	log.Debugf("[Deck Migration] Generated auth URL for server: %s", m.ServerURL)
	return authURL
}

// exchangeToken exchanges the OAuth authorization code for an access token
func (m *Migration) exchangeToken() (string, error) {
	serverURL := strings.TrimRight(m.ServerURL, "/")
	tokenURL := fmt.Sprintf("%s/index.php/apps/oauth2/api/v1/token", serverURL)

	// Get OAuth configuration
	clientID := config.MigrationDeckClientID.GetString()
	clientSecret := config.MigrationDeckClientSecret.GetString()
	redirectURL := config.MigrationDeckRedirectURL.GetString()

	if clientID == "" || clientSecret == "" || redirectURL == "" {
		return "", fmt.Errorf("OAuth not configured: missing client credentials")
	}

	// Prepare token request
	formData := url.Values{
		"client_id":     {clientID},
		"client_secret": {clientSecret},
		"code":          {m.Code},
		"redirect_uri":  {redirectURL},
		"grant_type":    {"authorization_code"},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", tokenURL, strings.NewReader(formData.Encode()))
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %w", err)
	}

	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	client := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return "", fmt.Errorf("token exchange request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("token exchange failed with status %d: %s", resp.StatusCode, string(body))
	}

	// Parse token response
	var tokenResp oauthTokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tokenResp); err != nil {
		return "", fmt.Errorf("failed to parse token response: %w", err)
	}

	if tokenResp.AccessToken == "" {
		return "", fmt.Errorf("received empty access token")
	}

	if strings.ToLower(tokenResp.TokenType) != "bearer" {
		return "", fmt.Errorf("unexpected token type: %s (expected Bearer)", tokenResp.TokenType)
	}

	log.Debugf("[Deck Migration] Successfully exchanged authorization code for access token")
	return tokenResp.AccessToken, nil
}

// Migrate performs the migration from Nextcloud Deck to Vikunja
// @Summary Migrate from Nextcloud Deck
// @Description Migrates all boards, stacks, cards, labels, comments, and attachments from Nextcloud Deck
// @tags migration
// @Accept json
// @Produce json
// @Security JWTKeyAuth
// @Param migrationData body deck.Migration true "Nextcloud server credentials"
// @Success 200 {object} models.Message "Migration successful"
// @Failure 400 {object} models.Message "Bad request"
// @Failure 401 {object} models.Message "Authentication failed"
// @Failure 500 {object} models.Message "Internal server error"
// @Router /migration/deck/migrate [post]
func (m *Migration) Migrate(u *user.User) error {
	log.Debugf("[Deck Migration] Starting migration for user %d from server %s", u.ID, m.ServerURL)

	// Exchange authorization code for access token
	accessToken, err := m.exchangeToken()
	if err != nil {
		return fmt.Errorf("failed to obtain access token: %w", err)
	}

	// Fetch boards using OAuth token
	boards, err := m.fetchBoards(accessToken)
	if err != nil {
		return fmt.Errorf("failed to fetch boards: %w", err)
	}

	log.Debugf("[Deck Migration] Found %d boards for user %d", len(boards), u.ID)

	// Build project structure
	var vikunjaProjects []*models.ProjectWithTasksAndBuckets

	for _, board := range boards {
		// Skip deleted boards
		if board.DeletedAt > 0 {
			log.Debugf("[Deck Migration] Skipping deleted board %d (%s)", board.ID, board.Title)
			continue
		}

		log.Debugf("[Deck Migration] Processing board %d: %s", board.ID, board.Title)

		project, err := m.convertBoardWithRetry(board, accessToken, u)
		if err != nil {
			log.Errorf("[Deck Migration] Failed to convert board %d after retries: %v", board.ID, err)
			continue
		}

		vikunjaProjects = append(vikunjaProjects, project)
	}

	// Resolve Nextcloud usernames to Vikunja users before insertion
	if len(vikunjaProjects) > 0 {
		err = m.resolveUserMappings(vikunjaProjects, u)
		if err != nil {
			log.Warningf("[Deck Migration] Failed to resolve some user mappings: %v", err)
			// Continue anyway - mappings that failed will use the migrating user
		}
	}

	// Insert all projects into Vikunja
	if len(vikunjaProjects) > 0 {
		err = migration.InsertFromStructure(vikunjaProjects, u)
		if err != nil {
			return fmt.Errorf("failed to insert projects: %w", err)
		}
	}

	log.Debugf("[Deck Migration] Migration complete for user %d: %d boards migrated", u.ID, len(vikunjaProjects))
	return nil
}

// getUserMapping loads and caches the user mapping from config
// Returns a map of Nextcloud username -> Vikunja user
func (m *Migration) getUserMapping(s *xorm.Session) map[string]*user.User {
	// Return cached mapping if already loaded
	if m.userMappingCache != nil {
		return m.userMappingCache
	}

	// Load mapping from config
	configMapping := config.MigrationDeckUserMapping.Get()
	if configMapping == nil {
		// No mapping configured, return empty map
		m.userMappingCache = make(map[string]*user.User)
		return m.userMappingCache
	}

	// Convert config map to user objects
	mapping := make(map[string]*user.User)

	// The config returns map[string]interface{}, we need to convert it
	for ncUsername, vikunjaUsernameInterface := range configMapping.(map[string]interface{}) {
		vikunjaUsername, ok := vikunjaUsernameInterface.(string)
		if !ok {
			log.Warningf("[Deck Migration] Invalid mapping for Nextcloud user '%s': value is not a string", ncUsername)
			continue
		}

		// Look up the Vikunja user
		vikunjaUser, err := user.GetUserByUsername(s, vikunjaUsername)
		if err != nil {
			log.Warningf("[Deck Migration] Failed to find Vikunja user '%s' for Nextcloud user '%s': %v", vikunjaUsername, ncUsername, err)
			continue
		}

		mapping[ncUsername] = vikunjaUser
		log.Debugf("[Deck Migration] Mapped Nextcloud user '%s' -> Vikunja user '%s' (ID: %d)", ncUsername, vikunjaUsername, vikunjaUser.ID)
	}

	m.userMappingCache = mapping
	log.Debugf("[Deck Migration] Loaded %d user mappings from config", len(mapping))
	return mapping
}

// mapUser looks up a Vikunja user from a Nextcloud username using the configured mapping
// Returns the mapped user if found, otherwise returns the fallbackUser
func (m *Migration) mapUser(s *xorm.Session, nextcloudUsername string, fallbackUser *user.User) *user.User {
	if nextcloudUsername == "" {
		return fallbackUser
	}

	// Load mapping (cached after first call)
	mapping := m.getUserMapping(s)

	// Look up user in mapping
	if vikunjaUser, exists := mapping[nextcloudUsername]; exists {
		log.Debugf("[Deck Migration] Mapped Nextcloud user '%s' -> Vikunja user '%s' (ID: %d)", nextcloudUsername, vikunjaUser.Username, vikunjaUser.ID)
		return vikunjaUser
	}

	// No mapping found, use fallback
	log.Debugf("[Deck Migration] No mapping found for Nextcloud user '%s', using fallback user (ID: %d)", nextcloudUsername, fallbackUser.ID)
	return fallbackUser
}

// resolveUserMappings resolves Nextcloud usernames stored in comments and attachments to Vikunja users
// This is called after building the project structure but before insertion
func (m *Migration) resolveUserMappings(projects []*models.ProjectWithTasksAndBuckets, fallbackUser *user.User) error {
	// Create a database session for user lookups
	s := db.NewSession()
	defer s.Close()

	commentsMapped := 0
	commentsTotal := 0
	attachmentsMapped := 0
	attachmentsTotal := 0

	for _, project := range projects {
		for _, task := range project.Tasks {
			// Resolve comment authors
			for _, comment := range task.Comments {
				commentsTotal++
				if comment.Author != nil && comment.Author.Username != "" && comment.Author.ID == 0 {
					// This is a temporary user object with Nextcloud username
					ncUsername := comment.Author.Username
					vikunjaUser := m.mapUser(s, ncUsername, fallbackUser)

					// Replace with resolved user
					comment.Author = vikunjaUser
					comment.AuthorID = vikunjaUser.ID

					if vikunjaUser.ID != fallbackUser.ID {
						commentsMapped++
					}
				}
			}

			// Resolve attachment creators
			for _, attachment := range task.Attachments {
				attachmentsTotal++
				if attachment.CreatedBy != nil && attachment.CreatedBy.Username != "" && attachment.CreatedBy.ID == 0 {
					// This is a temporary user object with Nextcloud username
					ncUsername := attachment.CreatedBy.Username
					vikunjaUser := m.mapUser(s, ncUsername, fallbackUser)

					// Replace with resolved user
					attachment.CreatedBy = vikunjaUser
					attachment.CreatedByID = vikunjaUser.ID

					if vikunjaUser.ID != fallbackUser.ID {
						attachmentsMapped++
					}
				}
			}
		}
	}

	log.Infof("[Deck Migration] User mapping summary: %d/%d comments mapped, %d/%d attachments mapped",
		commentsMapped, commentsTotal, attachmentsMapped, attachmentsTotal)

	return nil
}

// Helper functions

// nextcloudStatus represents the response from Nextcloud's status.php endpoint
type nextcloudStatus struct {
	Installed       bool   `json:"installed"`
	Version         string `json:"version"`
	VersionString   string `json:"versionstring"`
	Edition         string `json:"edition"`
	ProductName     string `json:"productname"`
	ExtendedSupport bool   `json:"extendedSupport"`
}

// validateNextcloudServer validates that the provided URL points to a reachable Nextcloud instance
func validateNextcloudServer(serverURL string) error {
	// Parse and validate URL format
	u, err := url.Parse(serverURL)
	if err != nil {
		return fmt.Errorf("invalid server URL format: %w", err)
	}

	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("invalid URL scheme: must be http or https")
	}

	// Enforce HTTPS for non-localhost addresses
	if u.Scheme == "http" {
		host := strings.ToLower(u.Hostname())
		if host != "localhost" && host != "127.0.0.1" && !strings.HasPrefix(host, "192.168.") && !strings.HasPrefix(host, "10.") {
			return fmt.Errorf("HTTPS is required for non-localhost servers")
		}
	}

	// Check if server is reachable and is a Nextcloud instance
	statusURL := strings.TrimRight(serverURL, "/") + "/status.php"
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", statusURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create validation request: %w", err)
	}

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("server unreachable: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("server returned status %d, expected 200", resp.StatusCode)
	}

	// Parse and validate Nextcloud status response
	var status nextcloudStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return fmt.Errorf("invalid response from server (not a Nextcloud instance): %w", err)
	}

	if !status.Installed {
		return fmt.Errorf("nextcloud is not properly installed on this server")
	}

	log.Debugf("[Deck Migration] Validated Nextcloud server: %s (version %s)", serverURL, status.VersionString)
	return nil
}

// generateOAuthState generates a cryptographically secure random state parameter for CSRF protection
func generateOAuthState() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", fmt.Errorf("failed to generate random state: %w", err)
	}
	return base64.URLEncoding.EncodeToString(b), nil
}

// truncateString truncates a string to maxLength characters
func truncateString(s string, maxLength int) string {
	if len(s) <= maxLength {
		return s
	}
	return s[:maxLength] + "..."
}

// makeRequest performs an HTTP request to the Deck API with OAuth Bearer token authentication
// If useOCS is true, uses the OCS API endpoint (/ocs/v2.php/apps/deck/api/v1.0/)
// Otherwise uses the regular API endpoint (/index.php/apps/deck/api/v1.0/)
func (m *Migration) makeRequest(endpoint, accessToken string) (*http.Response, error) {
	return m.makeRequestWithType(endpoint, accessToken, false)
}

// makeRequestWithType performs an HTTP request to the Deck API, allowing choice of endpoint type
// Includes retry logic with exponential backoff (max 15 seconds total)
func (m *Migration) makeRequestWithType(endpoint, accessToken string, useOCS bool) (*http.Response, error) {
	// Clean up server URL
	serverURL := strings.TrimRight(m.ServerURL, "/")

	var url string
	if useOCS {
		// OCS API endpoint: /ocs/v2.php/apps/deck/api/v1.0/
		url = fmt.Sprintf("%s/ocs/v2.php/apps/deck/api/%s%s", serverURL, apiVersion, endpoint)
	} else {
		// Regular API endpoint: /index.php/apps/deck/api/v1.0/
		url = fmt.Sprintf("%s/index.php/apps/deck/api/%s%s", serverURL, apiVersion, endpoint)
	}

	log.Debugf("[Deck Migration] Making API request to %s (useOCS=%v)", url, useOCS)

	return m.retryRequest(url, accessToken, 0)
}

// retryRequest performs an HTTP request with exponential backoff retry logic
// maxTotalDuration is 15 seconds, with initial backoff of 500ms doubling each retry
func (m *Migration) retryRequest(url, accessToken string, attempt int) (*http.Response, error) {
	return m.retryRequestWithHeaders(url, accessToken, attempt, nil)
}

// retryRequestWithHeaders performs an HTTP request with exponential backoff retry logic and custom headers
// Handles transient failures (context cancellation, timeouts, connection errors) with retries
// Request timeout is 60 seconds to allow slow Nextcloud instances to respond
// Network timeout for the HTTP client is 120 seconds as a fallback
func (m *Migration) retryRequestWithHeaders(url, accessToken string, attempt int, customHeaders map[string]string) (*http.Response, error) {
	const maxAttempts = 6
	const initialBackoff = 500 * time.Millisecond
	const requestTimeout = 60 * time.Second
	const clientTimeout = 120 * time.Second

	if attempt > 0 {
		// Calculate exponential backoff: 500ms, 1s, 2s, 4s, 8s
		backoff := initialBackoff * time.Duration(1<<(attempt-1))
		log.Debugf("[Deck Migration] Retrying request (attempt %d/%d) after %v", attempt, maxAttempts, backoff)
		time.Sleep(backoff)
	}

	ctx, cancel := context.WithTimeout(context.Background(), requestTimeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set required headers for Nextcloud Deck API with OAuth Bearer token
	req.Header.Set("OCS-APIRequest", "true")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Authorization", "Bearer "+accessToken)

	// Add custom headers if provided
	for key, value := range customHeaders {
		req.Header.Set(key, value)
	}

	client := &http.Client{
		Timeout: clientTimeout,
	}

	resp, err := client.Do(req)
	if err != nil {
		// Retry on network errors (connection refused, timeout, context canceled, etc.)
		if attempt < maxAttempts-1 {
			log.Warningf("[Deck Migration] Request failed (attempt %d/%d): %v - will retry", attempt+1, maxAttempts, err)
			return m.retryRequestWithHeaders(url, accessToken, attempt+1, customHeaders)
		}
		return nil, fmt.Errorf("request failed after %d attempts: %w", maxAttempts, err)
	}

	// Check for authentication errors (don't retry)
	if resp.StatusCode == 401 {
		resp.Body.Close()
		return nil, fmt.Errorf("authentication failed: invalid or expired access token")
	}

	// Retry on server errors (5xx)
	if resp.StatusCode >= 500 && resp.StatusCode < 600 && attempt < maxAttempts-1 {
		resp.Body.Close()
		log.Warningf("[Deck Migration] Server error %d (attempt %d/%d) - will retry", resp.StatusCode, attempt+1, maxAttempts)
		return m.retryRequestWithHeaders(url, accessToken, attempt+1, customHeaders)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("API returned status %d: %s", resp.StatusCode, string(body))
	}

	return resp, nil
}

// convertColor converts Deck color format to Vikunja hex color format
func convertColor(deckColor string) string {
	if deckColor == "" {
		return ""
	}
	// Deck colors are 6-character hex without #, Vikunja expects #RRGGBB
	if len(deckColor) == 6 {
		return "#" + deckColor
	}
	return ""
}

// parseDeckDate parses Deck API date strings (RFC3339 format)
func parseDeckDate(dateStr string) (time.Time, error) {
	if dateStr == "" || dateStr == "null" {
		return time.Time{}, nil
	}
	return time.Parse(time.RFC3339, dateStr)
}

// convertMarkdownToHTML converts markdown text to HTML to preserve formatting and newlines
// This ensures that newlines and other formatting from the source are properly preserved
func convertMarkdownToHTML(input string) (string, error) {
	if input == "" {
		return "", nil
	}
	var buf bytes.Buffer
	err := goldmark.Convert([]byte(input), &buf)
	if err != nil {
		return "", err
	}
	// #nosec - we are not responsible to escape this as we don't know the context where it is used
	return buf.String(), nil
}

// API fetch methods

func (m *Migration) fetchBoards(accessToken string) ([]*deckBoard, error) {
	resp, err := m.makeRequest("/boards", accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var boards []*deckBoard
	if err := json.NewDecoder(resp.Body).Decode(&boards); err != nil {
		return nil, fmt.Errorf("failed to decode boards: %w", err)
	}

	return boards, nil
}

func (m *Migration) fetchStacks(boardID int, accessToken string) ([]*deckStack, error) {
	resp, err := m.makeRequest(fmt.Sprintf("/boards/%d/stacks", boardID), accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var stacks []*deckStack
	if err := json.NewDecoder(resp.Body).Decode(&stacks); err != nil {
		return nil, fmt.Errorf("failed to decode stacks: %w", err)
	}

	return stacks, nil
}

func (m *Migration) fetchStackWithCards(boardID, stackID int, accessToken string) (*deckStack, error) {
	resp, err := m.makeRequest(fmt.Sprintf("/boards/%d/stacks/%d", boardID, stackID), accessToken)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read the body into a buffer to examine raw JSON
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	log.Debugf("[Deck Migration] Raw JSON response for stack %d/%d (first 1000 chars): %s",
		boardID, stackID, truncateString(string(bodyBytes), 1000))

	var stack deckStack
	if err := json.Unmarshal(bodyBytes, &stack); err != nil {
		return nil, fmt.Errorf("failed to decode stack with cards: %w", err)
	}

	// Log card descriptions from this stack
	log.Debugf("[Deck Migration] Stack %d has %d cards", stackID, len(stack.Cards))
	for i, card := range stack.Cards {
		if card.Description != "" {
			log.Debugf("[Deck Migration] Stack %d - Card %d/%d (%s) description length: %d, content: %q",
				stackID, i+1, len(stack.Cards), card.Title, len(card.Description), truncateString(card.Description, 300))
		}
	}

	return &stack, nil
}

func (m *Migration) fetchComments(cardID int, accessToken string) ([]*deckComment, error) {
	log.Debugf("[Deck Migration] Fetching comments for card %d using OCS API", cardID)
	resp, err := m.makeRequestWithType(fmt.Sprintf("/cards/%d/comments", cardID), accessToken, true)
	if err != nil {
		log.Warningf("[Deck Migration] Failed to fetch comments for card %d: %v", cardID, err)
		return nil, nil // Return empty slice, don't fail migration
	}
	defer resp.Body.Close()

	// Read the full response body to debug
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warningf("[Deck Migration] Failed to read response body for card %d: %v", cardID, err)
		return nil, nil
	}

	log.Debugf("[Deck Migration] Raw response for card %d (first 500 chars): %s", cardID, truncateString(string(bodyBytes), 500))

	// Parse OCS response wrapper
	var ocsResp ocsResponse
	if err := json.Unmarshal(bodyBytes, &ocsResp); err != nil {
		log.Warningf("[Deck Migration] Failed to decode OCS response for card %d: %v. First 200 chars: %s", cardID, err, truncateString(string(bodyBytes), 200))
		return nil, nil
	}

	log.Debugf("[Deck Migration] OCS response for card %d: status=%s, statusCode=%d", cardID, ocsResp.OCS.Meta.Status, ocsResp.OCS.Meta.StatusCode)

	// Extract and parse the data field
	var comments []*deckComment
	if ocsResp.OCS != nil && ocsResp.OCS.Data != nil {
		log.Debugf("[Deck Migration] Parsing comments data for card %d (raw length: %d bytes)", cardID, len(ocsResp.OCS.Data))
		if err := json.Unmarshal(ocsResp.OCS.Data, &comments); err != nil {
			log.Warningf("[Deck Migration] Failed to parse comments data for card %d: %v. Raw data: %s", cardID, err, string(ocsResp.OCS.Data))
			return nil, nil
		}
	} else {
		log.Debugf("[Deck Migration] No data in OCS response for card %d", cardID)
	}

	log.Debugf("[Deck Migration] Successfully fetched %d comments for card %d", len(comments), cardID)
	return comments, nil
}

func (m *Migration) fetchAttachments(boardID, stackID, cardID int, accessToken string) ([]*deckAttachment, error) {
	endpoint := fmt.Sprintf("/boards/%d/stacks/%d/cards/%d/attachments", boardID, stackID, cardID)
	log.Debugf("[Deck Migration] Making API request to fetch attachments for card %d: %s", cardID, endpoint)

	resp, err := m.makeRequest(endpoint, accessToken)
	if err != nil {
		log.Warningf("[Deck Migration] Failed to fetch attachments for card %d using documented endpoint: %v", cardID, err)
		return nil, nil
	}
	defer resp.Body.Close()

	log.Debugf("[Deck Migration] Received response for attachments endpoint (status: %d)", resp.StatusCode)

	// Read the full response body to debug
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warningf("[Deck Migration] Failed to read response body for card %d: %v", cardID, err)
		return nil, nil
	}

	log.Debugf("[Deck Migration] Raw response for card %d attachments (first 1000 chars): %s", cardID, truncateString(string(bodyBytes), 1000))
	log.Debugf("[Deck Migration] Raw response for card %d attachments body length: %d bytes", cardID, len(bodyBytes))

	var attachments []*deckAttachment
	if err := json.Unmarshal(bodyBytes, &attachments); err != nil {
		log.Warningf("[Deck Migration] Failed to decode attachments for card %d using documented endpoint: %v. Response (first 500 chars): %s", cardID, err, truncateString(string(bodyBytes), 500))
		log.Infof("[Deck Migration] Trying alternative undocumented endpoint for card %d attachments", cardID)
		return m.fetchAttachmentsAlternative(cardID, accessToken)
	}

	if len(attachments) > 0 {
		log.Infof("[Deck Migration] Successfully fetched %d attachments for card %d from documented endpoint %s", len(attachments), cardID, endpoint)
		for i, att := range attachments {
			log.Debugf("[Deck Migration] Attachment %d/%d: ID=%d, name=%s, type=%s, size=%d, deleted=%t", i+1, len(attachments), att.ID, att.Data, att.ExtendedData.Mimetype, att.ExtendedData.Filesize, att.DeletedAt > 0)
		}
		return attachments, nil
	}

	log.Debugf("[Deck Migration] No attachments found for card %d using documented endpoint (raw response was %d bytes)", cardID, len(bodyBytes))
	if len(bodyBytes) > 0 {
		log.Debugf("[Deck Migration] Full response for empty attachments on card %d: %s", cardID, string(bodyBytes))
	}

	// Try alternative endpoint if documented endpoint returns empty list
	log.Infof("[Deck Migration] Documented endpoint returned empty list for card %d, trying alternative undocumented endpoint", cardID)
	return m.fetchAttachmentsAlternative(cardID, accessToken)
}

// fetchAttachmentsAlternative tries to fetch attachments using the undocumented endpoint
// GET /apps/deck/cards/{cardId}/attachments
// This endpoint is used by Nextcloud Deck web UI and includes CORS checks
func (m *Migration) fetchAttachmentsAlternative(cardID int, accessToken string) ([]*deckAttachment, error) {
	serverURL := strings.TrimRight(m.ServerURL, "/")
	// Use the undocumented endpoint format: /apps/deck/cards/{cardId}/attachments
	url := fmt.Sprintf("%s/apps/deck/cards/%d/attachments", serverURL, cardID)
	log.Debugf("[Deck Migration] Making alternative API request for attachments on card %d: %s", cardID, url)

	// Add Origin header for CORS
	customHeaders := map[string]string{
		"Origin": serverURL,
	}

	resp, err := m.retryRequestWithHeaders(url, accessToken, 0, customHeaders)
	if err != nil {
		log.Warningf("[Deck Migration] Failed to fetch attachments for card %d using alternative endpoint: %v", cardID, err)
		return nil, nil
	}
	defer resp.Body.Close()

	log.Debugf("[Deck Migration] Received response for alternative attachments endpoint (status: %d)", resp.StatusCode)

	// Read the full response body
	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Warningf("[Deck Migration] Failed to read response body from alternative endpoint for card %d: %v", cardID, err)
		return nil, nil
	}

	log.Debugf("[Deck Migration] Raw response from alternative endpoint for card %d (first 1000 chars): %s", cardID, truncateString(string(bodyBytes), 1000))
	log.Debugf("[Deck Migration] Raw response from alternative endpoint for card %d body length: %d bytes", cardID, len(bodyBytes))

	var attachments []*deckAttachment
	if err := json.Unmarshal(bodyBytes, &attachments); err != nil {
		log.Warningf("[Deck Migration] Failed to decode attachments from alternative endpoint for card %d: %v. Response (first 500 chars): %s", cardID, err, truncateString(string(bodyBytes), 500))
		return nil, nil
	}

	if len(attachments) > 0 {
		log.Infof("[Deck Migration] Successfully fetched %d attachments for card %d from alternative endpoint", len(attachments), cardID)
		for i, att := range attachments {
			log.Debugf("[Deck Migration] Attachment %d/%d: ID=%d, name=%s, type=%s, fileID=%d, size=%d, deleted=%t", i+1, len(attachments), att.ID, att.Data, att.ExtendedData.Mimetype, att.ExtendedData.FileID, att.ExtendedData.Filesize, att.DeletedAt > 0)
		}
		return attachments, nil
	}

	log.Debugf("[Deck Migration] No attachments found for card %d using alternative endpoint (raw response was %d bytes)", cardID, len(bodyBytes))
	if len(bodyBytes) > 0 {
		log.Debugf("[Deck Migration] Full response from alternative endpoint for card %d: %s", cardID, string(bodyBytes))
	}

	return attachments, nil
}

// downloadAttachment attempts to download an attachment using multiple strategies
// 1. First tries Nextcloud Files API (what the web UI uses)
// 2. Falls back to Deck API endpoint if available
func (m *Migration) downloadAttachment(attachment *deckAttachment, boardID, stackID, cardID int, accessToken string) ([]byte, error) {
	log.Infof("[Deck Migration] Starting download of attachment %d (%s) - Expected size: %d bytes, MIME: %s",
		attachment.ID, attachment.Data, attachment.ExtendedData.Filesize, attachment.ExtendedData.Mimetype)

	// Try Nextcloud WebDAV first (what the web UI actually uses for downloads)
	// Uses: /remote.php/dav/files/{username}/{filePath}
	if attachment.ExtendedData.FileID > 0 && attachment.ExtendedData.Path != "" && attachment.ExtendedData.AttachmentCreator.ID != "" {
		log.Infof("[Deck Migration] Attempting download via Nextcloud WebDAV - fileID: %d, path: %s, user: %s",
			attachment.ExtendedData.FileID, attachment.ExtendedData.Path, attachment.ExtendedData.AttachmentCreator.ID)
		content, err := m.downloadViaNextcloudFiles(attachment, accessToken)
		if err == nil && len(content) > 0 {
			log.Infof("[Deck Migration] Successfully downloaded attachment %d via Nextcloud WebDAV (%d bytes)", attachment.ID, len(content))

			// Verify downloaded size
			if len(content) == int(attachment.ExtendedData.Filesize) {
				log.Debugf("[Deck Migration] ✓ Size verification PASSED: downloaded %d bytes matches expected %d bytes", len(content), attachment.ExtendedData.Filesize)
			} else {
				log.Warningf("[Deck Migration] ⚠ Size verification WARNING: downloaded %d bytes but expected %d bytes (diff: %d)", len(content), attachment.ExtendedData.Filesize, len(content)-int(attachment.ExtendedData.Filesize))
			}

			return content, nil
		}
		if err != nil {
			log.Warningf("[Deck Migration] Failed to download attachment %d via WebDAV: %v, trying Deck API", attachment.ID, err)
		} else {
			log.Warningf("[Deck Migration] WebDAV returned empty content for attachment %d, trying Deck API", attachment.ID)
		}
	} else {
		log.Debugf("[Deck Migration] Missing fileID, path, or creator ID for attachment %d, using Deck API endpoint", attachment.ID)
	}

	// Fallback to Deck API endpoint
	// Per Deck API docs: GET /boards/{boardId}/stacks/{stackId}/cards/{cardId}/attachments/{attachmentId}
	log.Infof("[Deck Migration] Downloading attachment %d (%s) using Deck API endpoint (board:%d, stack:%d, card:%d)",
		attachment.ID, attachment.Data, boardID, stackID, cardID)
	endpoint := fmt.Sprintf("/boards/%d/stacks/%d/cards/%d/attachments/%d", boardID, stackID, cardID, attachment.ID)
	log.Infof("[Deck Migration] FULL DECK API ENDPOINT: /boards/%d/stacks/%d/cards/%d/attachments/%d", boardID, stackID, cardID, attachment.ID)
	log.Debugf("[Deck Migration] Attachment metadata - FileID: %d, Path: %s, Size: %d bytes", attachment.ExtendedData.FileID, attachment.ExtendedData.Path, attachment.ExtendedData.Filesize)

	resp, err := m.makeRequest(endpoint, accessToken)
	if err != nil {
		log.Errorf("[Deck Migration] Failed to download attachment %d using Deck API: %v", attachment.ID, err)
		return nil, err
	}
	defer resp.Body.Close()

	log.Debugf("[Deck Migration] Received response for attachment %d download via Deck API (status: %d)", attachment.ID, resp.StatusCode)
	log.Debugf("[Deck Migration] Response headers - Content-Type: %s, Content-Length: %s, Content-Disposition: %s",
		resp.Header.Get("Content-Type"), resp.Header.Get("Content-Length"), resp.Header.Get("Content-Disposition"))

	buf := &bytes.Buffer{}
	bytesWritten, err := io.Copy(buf, resp.Body)
	if err != nil {
		log.Errorf("[Deck Migration] Failed to read attachment %d content from Deck API: %v", attachment.ID, err)
		return nil, fmt.Errorf("failed to download attachment: %w", err)
	}

	content := buf.Bytes()
	log.Infof("[Deck Migration] Successfully downloaded attachment %d via Deck API endpoint (%d bytes written, %d bytes total)", attachment.ID, bytesWritten, len(content))

	if len(content) != int(attachment.ExtendedData.Filesize) {
		log.Warningf("[Deck Migration] ⚠ Size verification WARNING: downloaded %d bytes but expected %d bytes (diff: %d)", len(content), attachment.ExtendedData.Filesize, len(content)-int(attachment.ExtendedData.Filesize))
	}

	return content, nil
}

// downloadViaNextcloudFiles downloads a file using the Nextcloud WebDAV endpoint
// The actual download uses: /remote.php/dav/files/{username}/{filePath}
// This is extracted from what the Nextcloud web UI does
func (m *Migration) downloadViaNextcloudFiles(attachment *deckAttachment, accessToken string) ([]byte, error) {
	serverURL := strings.TrimRight(m.ServerURL, "/")

	// Get the username from the attachment creator, fallback to a generic user
	username := attachment.ExtendedData.AttachmentCreator.ID
	if username == "" {
		log.Warningf("[Deck Migration] No creator ID found for attachment %d, download may fail", attachment.ID)
		return nil, fmt.Errorf("no attachment creator ID available")
	}

	// Path format from API: /Deck/filename.ext
	filePath := attachment.ExtendedData.Path

	// Construct the WebDAV URL: /remote.php/dav/files/{username}/{filePath}
	webdavURL := fmt.Sprintf("%s/remote.php/dav/files/%s%s", serverURL, username, filePath)
	log.Infof("[Deck Migration] Nextcloud WebDAV URL for attachment %d: /remote.php/dav/files/%s%s",
		attachment.ID, username, filePath)

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", webdavURL, nil)
	if err != nil {
		log.Errorf("[Deck Migration] Failed to create WebDAV request for attachment %d: %v", attachment.ID, err)
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", "Bearer "+accessToken)
	log.Debugf("[Deck Migration] Request headers for WebDAV - Authorization: Bearer***")

	client := &http.Client{
		Timeout: 180 * time.Second,
	}

	resp, err := client.Do(req)
	if err != nil {
		log.Errorf("[Deck Migration] Failed to download attachment %d via WebDAV: %v", attachment.ID, err)
		return nil, fmt.Errorf("failed to download: %w", err)
	}
	defer resp.Body.Close()

	log.Debugf("[Deck Migration] Received response for attachment %d via WebDAV (status: %d, Content-Type: %s, Content-Length: %s)",
		attachment.ID, resp.StatusCode, resp.Header.Get("Content-Type"), resp.Header.Get("Content-Length"))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		log.Errorf("[Deck Migration] WebDAV returned status %d for attachment %d: %s", resp.StatusCode, attachment.ID, truncateString(string(body), 300))
		return nil, fmt.Errorf("webdav returned status %d", resp.StatusCode)
	}

	buf := &bytes.Buffer{}
	bytesWritten, err := io.Copy(buf, resp.Body)
	if err != nil {
		log.Errorf("[Deck Migration] Failed to read attachment %d content: %v", attachment.ID, err)
		return nil, fmt.Errorf("failed to read content: %w", err)
	}

	content := buf.Bytes()
	log.Infof("[Deck Migration] Downloaded attachment %d via WebDAV (%d bytes written, %d bytes total)", attachment.ID, bytesWritten, len(content))

	// Verification
	if len(content) == 0 {
		log.Warningf("[Deck Migration] WARNING: Downloaded attachment %d has 0 bytes", attachment.ID)
		return nil, fmt.Errorf("downloaded content is empty")
	}
	if bytesWritten != int64(len(content)) {
		log.Warningf("[Deck Migration] WARNING: Bytes written (%d) does not match buffer size (%d) for attachment %d", bytesWritten, len(content), attachment.ID)
	}

	return content, nil
}

// Conversion functions

// convertBoardWithRetry converts a board with retry logic on transient failures
// Retries up to 3 times with exponential backoff on context cancellation or timeout errors
func (m *Migration) convertBoardWithRetry(board *deckBoard, accessToken string, u *user.User) (*models.ProjectWithTasksAndBuckets, error) {
	const maxBoardRetries = 3
	const initialBoardBackoff = 1 * time.Second

	for attempt := 1; attempt <= maxBoardRetries; attempt++ {
		log.Debugf("[Deck Migration] Converting board %d (attempt %d/%d)", board.ID, attempt, maxBoardRetries)

		project, err := m.convertBoard(board, accessToken, u)
		if err == nil {
			return project, nil
		}

		// Check if this is a transient error worth retrying
		errStr := err.Error()
		isTransientError := strings.Contains(errStr, "context canceled") ||
			strings.Contains(errStr, "context deadline exceeded") ||
			strings.Contains(errStr, "timeout") ||
			strings.Contains(errStr, "connection refused") ||
			strings.Contains(errStr, "connection reset")

		if !isTransientError || attempt == maxBoardRetries {
			// Not a transient error or we've exhausted retries
			return nil, err
		}

		// Exponential backoff: 1s, 2s, 4s
		backoff := initialBoardBackoff * time.Duration(1<<(attempt-1))
		log.Warningf("[Deck Migration] Board %d conversion failed (attempt %d/%d): %v - retrying in %v", board.ID, attempt, maxBoardRetries, err, backoff)
		time.Sleep(backoff)
	}

	return nil, fmt.Errorf("board conversion failed after %d attempts", maxBoardRetries)
}

func (m *Migration) convertBoard(board *deckBoard, accessToken string, _ *user.User) (*models.ProjectWithTasksAndBuckets, error) {
	log.Infof("[Deck Migration] Converting board %d (%s) to Vikunja project", board.ID, board.Title)
	// Create label map
	labelMap := m.buildLabelMap(board.Labels)
	log.Debugf("[Deck Migration] Board %d has %d labels", board.ID, len(labelMap))

	// Fetch stacks for this board
	stacks, err := m.fetchStacks(board.ID, accessToken)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch stacks: %w", err)
	}
	log.Debugf("[Deck Migration] Board %d has %d stacks", board.ID, len(stacks))

	project := &models.ProjectWithTasksAndBuckets{
		Project: models.Project{
			Title:      board.Title,
			HexColor:   convertColor(board.Color),
			IsArchived: board.Archived,
		},
		Buckets: []*models.Bucket{},
		Tasks:   []*models.TaskWithComments{},
	}

	// Convert each stack to a bucket
	for _, stack := range stacks {
		if stack.DeletedAt > 0 {
			log.Debugf("[Deck Migration] Skipping deleted stack %d (%s)", stack.ID, stack.Title)
			continue
		}

		// Fetch full stack with cards
		fullStack, err := m.fetchStackWithCards(board.ID, stack.ID, accessToken)
		if err != nil {
			log.Errorf("[Deck Migration] Failed to fetch stack %d with cards: %v", stack.ID, err)
			continue
		}

		bucket := &models.Bucket{
			ID:    int64(stack.ID), // Temporary ID for mapping, will be reset by InsertFromStructure
			Title: stack.Title,
			// Position calculation: (order + 1) * 65536
			// - Add 1 to avoid position 0 (which Vikunja treats as "unset")
			// - Multiply by 65536 to use Vikunja's standard bucket spacing
			// - Preserves left-to-right order from Nextcloud Deck
			Position: float64(stack.Order+1) * 65536.0,
		}
		project.Buckets = append(project.Buckets, bucket)

		// Convert cards to tasks
		for _, card := range fullStack.Cards {
			if card.DeletedAt > 0 {
				log.Debugf("[Deck Migration] Skipping deleted card %d (%s)", card.ID, card.Title)
				continue
			}

			task := m.convertCard(&card, board.ID, stack.ID, labelMap, int64(stack.ID), accessToken)
			project.Tasks = append(project.Tasks, task)
		}
	}

	log.Infof("[Deck Migration] Completed conversion of board %d (%s): %d buckets, %d tasks total", board.ID, board.Title, len(project.Buckets), len(project.Tasks))
	return project, nil
}

func (m *Migration) buildLabelMap(deckLabels []deckLabel) map[int]*models.Label {
	labelMap := make(map[int]*models.Label)

	for _, dl := range deckLabels {
		// Skip deleted labels
		if dl.DeletedAt > 0 {
			continue
		}

		label := &models.Label{
			Title:    dl.Title,
			HexColor: convertColor(dl.Color),
		}
		labelMap[dl.ID] = label
	}

	return labelMap
}

func (m *Migration) convertCard(card *deckCard, boardID, stackID int, labelMap map[int]*models.Label, bucketID int64, accessToken string) *models.TaskWithComments {
	// Parse due date
	dueDate, err := parseDeckDate(card.DueDate)
	if err != nil {
		log.Debugf("[Deck Migration] Failed to parse due date for card %d: %v", card.ID, err)
	}

	// Log raw description from API with readable formatting
	if len(card.Description) > 0 {
		// Show line breaks - count newlines
		newlineCount := strings.Count(card.Description, "\n")
		log.Infof("[Deck Migration] Card %d (%s) - Description from API: %d chars, %d newlines",
			card.ID, card.Title, len(card.Description), newlineCount)

		// Log the actual content, making newlines visible
		descWithVisibleNewlines := strings.ReplaceAll(card.Description, "\n", "\\n")
		log.Debugf("[Deck Migration] Card %d - Content: %s", card.ID, truncateString(descWithVisibleNewlines, 500))

		// Check for different whitespace patterns
		hasHTMLBr := strings.Contains(card.Description, "<br")
		hasTabs := strings.Contains(card.Description, "\t")
		hasNonBreakingSpace := strings.Contains(card.Description, "\u00A0")
		log.Debugf("[Deck Migration] Card %d - Special chars: HTML_BR=%v, tabs=%v, nbsp=%v",
			card.ID, hasHTMLBr, hasTabs, hasNonBreakingSpace)
	} else {
		log.Debugf("[Deck Migration] Card %d (%s) - No description", card.ID, card.Title)
	}

	// Build description with assignees
	description := m.normalizeDescription(card.Description)

	if len(card.AssignedUsers) > 0 {
		description += m.formatAssignees(card.AssignedUsers)
	}

	// Log the final description that will be stored
	descNewlinesAfter := strings.Count(description, "\n")
	if len(description) > 0 {
		descWithVisibleNewlines := strings.ReplaceAll(description, "\n", "\\n")
		log.Debugf("[Deck Migration] Card %d - Final description for storage: %d chars, %d newlines",
			card.ID, len(description), descNewlinesAfter)
		log.Debugf("[Deck Migration] Card %d - Content preview: %s", card.ID, truncateString(descWithVisibleNewlines, 300))
	}

	// Determine if task is done
	done := card.Archived || (card.Done != nil && *card.Done != "")
	var doneAt time.Time
	if card.Done != nil && *card.Done != "" {
		doneAt, err = time.Parse(time.RFC3339, *card.Done)
		if err != nil {
			log.Debugf("[Deck Migration] Failed to parse done date for card %d: %v", card.ID, err)
		}
	}

	createdTime := time.Unix(card.CreatedAt, 0)
	updatedTime := time.Unix(card.LastModified, 0)

	log.Debugf("[Deck Migration] Card %d (%s): Setting Created=%s, Updated=%s",
		card.ID, card.Title, createdTime.Format(time.RFC3339), updatedTime.Format(time.RFC3339))

	task := &models.TaskWithComments{
		Task: models.Task{
			Title:       card.Title,
			Description: description,
			Done:        done,
			DoneAt:      doneAt,
			DueDate:     dueDate,
			Position:    float64(card.Order),
			Created:     createdTime,
			Updated:     updatedTime,
			BucketID:    bucketID,
		},
		Comments: []*models.TaskComment{},
	}

	// Verify the description is in the task object before it's returned
	if len(task.Description) > 0 {
		log.Infof("[Deck Migration] Card %d: Task object created with description (%d chars)", card.ID, len(task.Description))
	}

	// Assign labels (through the embedded Task struct)
	task.Labels = []*models.Label{}
	for _, deckLabel := range card.Labels {
		// Skip deleted labels
		if deckLabel.DeletedAt > 0 {
			continue
		}

		// Use the label from the map if it exists, otherwise create from card label
		if label, exists := labelMap[deckLabel.ID]; exists {
			task.Labels = append(task.Labels, label)
		} else {
			// Create label from the card's label data
			label := &models.Label{
				Title:    deckLabel.Title,
				HexColor: convertColor(deckLabel.Color),
			}
			task.Labels = append(task.Labels, label)
		}
	}

	// Initialize attachments slice
	task.Attachments = []*models.TaskAttachment{}

	// Fetch and convert comments
	log.Debugf("[Deck Migration] Processing comments for card %d (%s)", card.ID, card.Title)
	comments, err := m.fetchComments(card.ID, accessToken)
	if err != nil {
		log.Warningf("[Deck Migration] Error fetching comments for card %d: %v", card.ID, err)
	}
	if len(comments) > 0 {
		log.Debugf("[Deck Migration] Converting %d comments for card %d", len(comments), card.ID)
		for i, comment := range comments {
			log.Debugf("[Deck Migration] Converting comment %d/%d for card %d - Author: %s", i+1, len(comments), card.ID, comment.ActorDisplayName)
			taskComment := m.convertComment(comment)
			task.Comments = append(task.Comments, taskComment)
			log.Debugf("[Deck Migration] Successfully converted comment %d/%d", i+1, len(comments))
		}
		log.Infof("[Deck Migration] Successfully added %d comments to card %d (%s)", len(comments), card.ID, card.Title)
	} else {
		log.Debugf("[Deck Migration] No comments found for card %d (%s)", card.ID, card.Title)
	}

	// Fetch and convert attachments
	log.Debugf("[Deck Migration] Fetching attachments for card %d (%s) from board %d, stack %d", card.ID, card.Title, boardID, stackID)
	attachments, err := m.fetchAttachments(boardID, stackID, card.ID, accessToken)
	if err != nil {
		log.Warningf("[Deck Migration] Error fetching attachments for card %d: %v", card.ID, err)
	}
	if len(attachments) > 0 {
		log.Infof("[Deck Migration] ========== ATTACHMENT PROCESSING START ==========")
		log.Infof("[Deck Migration] Card %d (%s) - Processing %d attachments", card.ID, card.Title, len(attachments))
		attachmentsProcessed := 0
		attachmentsFailed := 0

		for i, attachment := range attachments {
			if attachment.DeletedAt > 0 {
				log.Debugf("[Deck Migration] Skipping deleted attachment %d for card %d", attachment.ID, card.ID)
				continue
			}

			log.Infof("[Deck Migration] [%d/%d] Processing: %s (ID: %d, FileID: %d, Size: %d bytes, Type: %s)",
				i+1, len(attachments), attachment.Data, attachment.ID, attachment.ExtendedData.FileID, attachment.ExtendedData.Filesize, attachment.ExtendedData.Mimetype)

			taskAttachment, err := m.convertAttachment(attachment, boardID, stackID, card.ID, accessToken)
			if err != nil {
				log.Errorf("[Deck Migration] ✗ FAILED: Attachment %d/%d (%s) conversion error: %v", i+1, len(attachments), attachment.Data, err)
				attachmentsFailed++
				continue
			}

			// Verify attachment before adding to task
			if taskAttachment == nil {
				log.Errorf("[Deck Migration] ✗ FAILED: Attachment %d/%d (%s) - returned nil TaskAttachment", i+1, len(attachments), attachment.Data)
				attachmentsFailed++
				continue
			}

			beforeCount := len(task.Attachments)
			task.Attachments = append(task.Attachments, taskAttachment)
			afterCount := len(task.Attachments)

			if afterCount != beforeCount+1 {
				log.Errorf("[Deck Migration] ✗ FAILED: Attachment %d/%d (%s) - failed to append to task (before: %d, after: %d)",
					i+1, len(attachments), attachment.Data, beforeCount, afterCount)
				attachmentsFailed++
				continue
			}

			log.Infof("[Deck Migration] ✓ SUCCESS: Attachment %d/%d (%s) - Appended to task.Attachments (now has %d attachments)",
				i+1, len(attachments), attachment.Data, afterCount)
			attachmentsProcessed++
		}

		log.Infof("[Deck Migration] ========== ATTACHMENT PROCESSING COMPLETE ==========")
		log.Infof("[Deck Migration] Card %d (%s) - Summary: %d processed, %d failed, %d total in task.Attachments",
			card.ID, card.Title, attachmentsProcessed, attachmentsFailed, len(task.Attachments))

		if len(task.Attachments) != attachmentsProcessed {
			log.Warningf("[Deck Migration] ⚠ WARNING: Attachments processed (%d) != task.Attachments length (%d)", attachmentsProcessed, len(task.Attachments))
		}
	} else {
		log.Debugf("[Deck Migration] No attachments found for card %d (%s)", card.ID, card.Title)
	}

	return task
}

func (m *Migration) normalizeDescription(desc string) string {
	if desc == "" {
		return ""
	}

	// First, normalize wrapped URLs and convert HTML breaks to preserve newlines
	// Remove angle brackets wrapping URLs/content (Nextcloud Deck wraps URLs like <https://example.com>)
	desc = normalizeAngleBrackets(desc)

	// Convert HTML line breaks to actual newlines for markdown processing
	desc = strings.ReplaceAll(desc, "<br>", "\n")
	desc = strings.ReplaceAll(desc, "<br/>", "\n")
	desc = strings.ReplaceAll(desc, "<br />", "\n")

	log.Debugf("[Deck Migration] Description before HTML conversion: %d chars", len(desc))

	// Convert markdown (with preserved newlines) to HTML to maintain formatting
	// This ensures newlines are preserved in the final output
	htmlDesc, err := convertMarkdownToHTML(desc)
	if err != nil {
		log.Warningf("[Deck Migration] Failed to convert description to HTML: %v, falling back to plain text", err)
		return desc
	}

	if htmlDesc != "" {
		log.Debugf("[Deck Migration] Description after HTML conversion: %d chars (was %d)", len(htmlDesc), len(desc))
		return htmlDesc
	}

	return desc
}

// normalizeAngleBrackets removes angle brackets that wrap URLs or other content
// This fixes <https://example.com> -> https://example.com
// But preserves actual HTML tags like <br>, <div>, etc.
func normalizeAngleBrackets(desc string) string {
	// Match pattern: <content> where content looks like a URL or non-HTML content
	// We use a regex to find <...> that don't look like HTML tags
	// HTML tags typically have letters followed by optional attributes/content
	// URLs start with http, ftp, or look like domain names

	// First, handle wrapped URLs: <https://...>, <http://...>, <ftp://...>
	desc = strings.NewReplacer(
		"<https://", "https://",
		"<http://", "http://",
		"<ftp://", "ftp://",
	).Replace(desc)

	// Now remove trailing > that was paired with the opening <
	// This is safe because we only removed specific opening brackets above
	// We need to find patterns like:  https://example.com> and remove the >
	// But we need to be careful not to remove > that's part of actual HTML

	// Use a regex to find and replace wrapped URLs with trailing >
	// Match: protocol://...anything... followed by >
	desc = replaceWrappedURLs(desc)

	return desc
}

// replaceWrappedURLs removes the trailing > from URLs that were wrapped in <>
func replaceWrappedURLs(desc string) string {
	// This regex matches URLs (with or without trailing >) that we just unwrapped
	// We match from http/https/ftp to the next > or space/newline
	var result strings.Builder
	i := 0
	for i < len(desc) {
		// Look for patterns like: https://... (without leading <)
		if i < len(desc)-1 && (strings.HasPrefix(desc[i:], "https://") ||
			strings.HasPrefix(desc[i:], "http://") ||
			strings.HasPrefix(desc[i:], "ftp://")) {

			// Find the end of the URL
			// URLs typically end at whitespace, newline, or >
			j := i
			for j < len(desc) && desc[j] != '>' && desc[j] != ' ' && desc[j] != '\n' && desc[j] != '\t' {
				j++
			}

			// Copy the URL without the trailing >
			result.WriteString(desc[i:j])

			// If there's a >, skip it
			if j < len(desc) && desc[j] == '>' {
				j++
			}

			i = j
		} else {
			result.WriteByte(desc[i])
			i++
		}
	}

	return result.String()
}

func (m *Migration) formatAssignees(assignedUsers []deckAssignedUser) string {
	if len(assignedUsers) == 0 {
		return ""
	}

	result := "\n\n---\n**Originally assigned to:**\n"
	for _, au := range assignedUsers {
		result += fmt.Sprintf("- @%s\n", au.Participant.DisplayName)
	}
	return result
}

func (m *Migration) convertComment(comment *deckComment) *models.TaskComment {
	log.Debugf("[Deck Migration] Converting comment ID %d from user %s (NC username: %s)", comment.ID, comment.ActorDisplayName, comment.ActorID)
	created, err := time.Parse(time.RFC3339, comment.CreationDateTime)
	if err != nil {
		log.Warningf("[Deck Migration] Failed to parse comment date for comment ID %d (%s): %v", comment.ID, comment.CreationDateTime, err)
		created = time.Now()
	}

	log.Debugf("[Deck Migration] Comment %d: Setting Created=%s (from: %s)",
		comment.ID, created.Format(time.RFC3339), comment.CreationDateTime)

	// Prepend author information to comment
	message := fmt.Sprintf("**Originally by: %s**\n\n%s", comment.ActorDisplayName, comment.Message)
	log.Debugf("[Deck Migration] Comment message prepared (length: %d bytes) for comment ID %d", len(message), comment.ID)

	taskComment := &models.TaskComment{
		Comment: message,
		Created: created,
		// Store Nextcloud username temporarily in Author field for later resolution
		// This will be resolved to actual Vikunja user in create_from_structure.go
		Author: &user.User{
			Username: comment.ActorID, // Nextcloud username
		},
	}
	log.Debugf("[Deck Migration] Comment ID %d converted successfully with Created=%s, NC author=%s", comment.ID, created.Format(time.RFC3339), comment.ActorID)
	return taskComment
}

func (m *Migration) convertAttachment(attachment *deckAttachment, boardID, stackID, cardID int, accessToken string) (*models.TaskAttachment, error) {
	log.Infof("[Deck Migration] CONVERSION START: attachment %d (%s)", attachment.ID, attachment.Data)
	log.Debugf("[Deck Migration]   Details - Type: %s, Size: %d bytes, CreatedAt: %s, FileID: %d, Path: %s",
		attachment.ExtendedData.Mimetype, attachment.ExtendedData.Filesize,
		time.Unix(attachment.CreatedAt, 0).Format(time.RFC3339), attachment.ExtendedData.FileID, attachment.ExtendedData.Path)

	// Download attachment content
	log.Debugf("[Deck Migration]   Step 1: Downloading content...")
	content, err := m.downloadAttachment(attachment, boardID, stackID, cardID, accessToken)
	if err != nil {
		log.Errorf("[Deck Migration] ✗ FAILED: Download failed for attachment %d (%s): %v", attachment.ID, attachment.Data, err)
		return nil, fmt.Errorf("failed to download attachment: %w", err)
	}

	log.Infof("[Deck Migration]   Step 1 COMPLETE: Downloaded %d bytes (expected %d bytes)", len(content), attachment.ExtendedData.Filesize)

	// Validate content
	log.Debugf("[Deck Migration]   Step 2: Validating downloaded content...")
	if len(content) == 0 {
		log.Errorf("[Deck Migration] ✗ VALIDATION FAILED: Attachment %d has empty content", attachment.ID)
		return nil, fmt.Errorf("downloaded attachment has 0 bytes")
	}
	log.Debugf("[Deck Migration]   Step 2 COMPLETE: Content validation passed (non-empty)")

	// Create TaskAttachment object
	log.Debugf("[Deck Migration]   Step 3: Creating TaskAttachment object...")
	createdTime := time.Unix(attachment.CreatedAt, 0)

	// Get Nextcloud username from attachment (prefer CreatedBy, fallback to AttachmentCreator.ID)
	ncUsername := attachment.CreatedBy
	if ncUsername == "" && attachment.ExtendedData.AttachmentCreator.ID != "" {
		ncUsername = attachment.ExtendedData.AttachmentCreator.ID
	}

	log.Debugf("[Deck Migration]     File.Name: %s", attachment.Data)
	log.Debugf("[Deck Migration]     File.Mime: %s", attachment.ExtendedData.Mimetype)
	log.Debugf("[Deck Migration]     File.Size: %d", uint64(attachment.ExtendedData.Filesize))
	log.Debugf("[Deck Migration]     File.Created: %s", createdTime.Format(time.RFC3339))
	log.Debugf("[Deck Migration]     FileContent length: %d bytes", len(content))
	log.Debugf("[Deck Migration]     NC Creator: %s", ncUsername)

	log.Debugf("[Deck Migration]     Attachment.Created: %s", createdTime.Format(time.RFC3339))
	log.Debugf("[Deck Migration]     File.Created: %s", createdTime.Format(time.RFC3339))

	taskAttachment := &models.TaskAttachment{
		File: &files.File{
			Name:        attachment.Data,
			Mime:        attachment.ExtendedData.Mimetype,
			Size:        uint64(attachment.ExtendedData.Filesize),
			FileContent: content,
			Created:     createdTime,
		},
		Created: createdTime,
		// Store Nextcloud username temporarily in CreatedBy field for later resolution
		// This will be resolved to actual Vikunja user in create_from_structure.go
		CreatedBy: &user.User{
			Username: ncUsername, // Nextcloud username
		},
	}

	// Verify file content was preserved
	if len(taskAttachment.File.FileContent) == 0 {
		log.Errorf("[Deck Migration] ✗ FAILED: TaskAttachment.File.FileContent is empty for attachment %d", attachment.ID)
		return nil, fmt.Errorf("FileContent lost during object creation")
	}

	log.Infof("[Deck Migration]   Step 3 COMPLETE: TaskAttachment object created successfully")
	log.Infof("[Deck Migration] ✓ CONVERSION SUCCESS: attachment %d (%s) - %d bytes ready for insertion", attachment.ID, attachment.Data, len(taskAttachment.File.FileContent))

	return taskAttachment, nil
}
