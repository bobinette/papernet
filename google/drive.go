package google

import (
	"bytes"
	"fmt"
	"net/http"

	"google.golang.org/api/drive/v3"
	"google.golang.org/api/googleapi"

	"github.com/bobinette/papernet/errors"
)

type GDriveService struct {
	service *drive.Service
}

func NewGDriveService(client *http.Client) (DriveService, error) {
	ds, err := drive.New(client)
	if err != nil {
		return nil, err
	}

	return &GDriveService{
		service: ds,
	}, nil
}

func (ds *GDriveService) UserHasAllowedDrive() (bool, error) {
	_, err := ds.service.Files.List().PageSize(1).
		Fields("nextPageToken, files(id, name)").Do()

	// No error means the user already has access to the drive
	if err == nil {
		return true, nil
	}

	isInsufficientPermission := false
	code := 500
	if err, ok := err.(*googleapi.Error); ok {
		code = err.Code
		for _, e := range err.Errors {
			if e.Reason == errInsufficientPermissions {
				isInsufficientPermission = true
				break
			}
		}
	}

	// The error is something else, returning to the caller
	if !isInsufficientPermission {
		return false, errors.New("unable to check permission: %v\n", errors.WithCause(err), errors.WithCode(code))
	}

	return false, nil
}

func (ds *GDriveService) ListFiles(folderID, name string) ([]DriveFile, string, error) {
	q := fmt.Sprintf("trashed = false and '%s' in parents", folderID)
	if name != "" {
		q = fmt.Sprintf("%s and name contains '%s'", q, name)
	}

	fmt.Println(q)

	r, err := ds.service.Files.
		List().
		Q(q).
		PageSize(10).
		Fields("nextPageToken, files(id, name, mimeType, webViewLink)").
		OrderBy("name").
		Do()
	if err != nil {
		if err, ok := err.(*googleapi.Error); ok {
			fmt.Printf("%+v\n", err.Body)
		}
		return nil, "", fmt.Errorf("unable to retrieve files: %v\n", err)
	}

	files := make([]DriveFile, len(r.Files))
	for i, file := range r.Files {
		files[i] = DriveFile{
			ID:       file.Id,
			Name:     file.Name,
			URL:      file.WebViewLink,
			MimeType: file.MimeType,
		}
	}
	return files, "", nil
}

func (ds *GDriveService) CreateFile(name, typ, folderID string, data []byte) (DriveFile, error) {
	file := &drive.File{
		Name:     name,
		MimeType: typ,
	}
	file.Parents = []string{folderID}

	reader := bytes.NewReader(data)
	createRes, err := ds.service.Files.
		Create(file).
		Media(reader).
		Fields("id, name, mimeType, webViewLink").
		Do()
	if err != nil {
		return DriveFile{}, errors.New("error upload file to Google Drive", errors.WithCause(err))
	}

	dFile := DriveFile{
		ID:       createRes.Id,
		Name:     createRes.Name,
		MimeType: createRes.MimeType,
		URL:      createRes.WebViewLink,
	}

	return dFile, nil
}

func (ds *GDriveService) GetFolderID(name string) (string, error) {
	// Adapted from https://gist.github.com/TheGU/e6d0ae13f2fa83f3bd8d
	q := fmt.Sprintf("name='%s' and mimeType='application/vnd.google-apps.folder'", name)
	filesRes, err := ds.service.Files.List().Q(q).PageSize(1).Do()
	if err != nil {
		return "", errors.New("unable to retrieve folder", errors.WithCause(err))
	}

	if len(filesRes.Files) > 0 {
		return filesRes.Files[0].Id, nil
	}

	return "", nil
}

func (ds *GDriveService) CreateFolder(name string) (string, error) {
	f := &drive.File{
		Name:        name,
		Description: "Auto Created by Papernet",
		MimeType:    "application/vnd.google-apps.folder",
	}
	createRes, err := ds.service.Files.Create(f).Do()
	if err != nil {
		return "", errors.New("could not create folder", errors.WithCause(err))
	}

	return createRes.Id, nil
}
