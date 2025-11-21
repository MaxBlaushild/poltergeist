package dropbox

import "time"

// File represents a Dropbox file or folder
type File struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	PathLower     string    `json:"pathLower"`
	PathDisplay   string    `json:"pathDisplay"`
	Size          int64     `json:"size"`
	ClientModified time.Time `json:"clientModified"`
	ServerModified time.Time `json:"serverModified"`
	Rev           string    `json:"rev"`
	ContentHash   string    `json:"contentHash,omitempty"`
	IsFolder      bool      `json:"isFolder"`
	SharedLink    string    `json:"sharedLink,omitempty"`
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// FileListResponse represents file list response
type FileListResponse struct {
	Files     []File `json:"files"`
	HasMore   bool   `json:"hasMore"`
	Cursor    string `json:"cursor,omitempty"`
}

