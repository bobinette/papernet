package google

import (
	"net/http"

	"golang.org/x/oauth2"
)

type User struct {
	ID       int    `json:"id"`
	GoogleID string `json:"googleId"`

	Tokens map[string]*oauth2.Token `json:"tokens"`
}

type UserRepository interface {
	GetByID(id int) (User, error)
	GetByGoogleID(googleID string) (User, error)

	Upsert(User) error
}

type DriveFile struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	URL      string `json:"url"`
	MimeType string `json:"mimeType"`
}

type DriveServiceFactory func(*http.Client) (DriveService, error)

type DriveService interface {
	UserHasAllowedDrive() (bool, error)
	GetFolderID(name string) (string, error)

	ListFiles(folderID string, pageToken string) ([]DriveFile, string, error)
	CreateFile(name, typ, folderID string, data []byte) (DriveFile, error)
	CreateFolder(name string) (string, error)
}
