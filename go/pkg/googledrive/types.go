package googledrive

import "time"

// File represents a Google Drive file
type File struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	MimeType      string    `json:"mimeType"`
	Size          int64     `json:"size,string"`
	CreatedTime   time.Time `json:"createdTime"`
	ModifiedTime  time.Time `json:"modifiedTime"`
	WebViewLink   string    `json:"webViewLink"`
	WebContentLink string   `json:"webContentLink"`
	Owners        []Owner   `json:"owners"`
	Shared        bool      `json:"shared"`
}

// Owner represents a file owner
type Owner struct {
	DisplayName string `json:"displayName"`
	EmailAddress string `json:"emailAddress"`
}

// Permission represents a file permission
type Permission struct {
	ID         string `json:"id"`
	Type       string `json:"type"` // user, group, domain, anyone
	Role       string `json:"role"` // reader, writer, commenter, owner
	EmailAddress string `json:"emailAddress,omitempty"`
}

// TokenResponse represents OAuth token response
type TokenResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresIn    int       `json:"expires_in"`
	ExpiresAt    time.Time `json:"expires_at"`
	TokenType    string    `json:"token_type"`
}

// FileListResponse represents paginated file list response
type FileListResponse struct {
	Files         []File `json:"files"`
	NextPageToken string `json:"nextPageToken"`
}

