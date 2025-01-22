package cmd

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"encoding/base64"
	"mime/multipart"
	"net/http"
)

func NewMultipartRequest(url string, params map[string]string, fpath string, auth string) (*http.Request, error) {
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	part, err := writer.CreateFormFile("file", filepath.Base(fpath))
	FatalError("newMultipartRequest", err)

	fp, err := os.Open(fpath)
	FatalError("newMultipartRequest", err)

	_, err = io.Copy(part, fp)

	for key, val := range params {
		_ = writer.WriteField(key, val)
	}
	err = writer.Close()
	if err != nil {
		return nil, err
	}

	request, err := http.NewRequest("POST", url, body)
	request.Header.Add("Content-Type", writer.FormDataContentType())
	request.Header.Add("Authorization", "Basic "+auth)
	return request, err
}

func clientPostFile(user, group, key, file, endpoint, auth string) error {

	extraParams := map[string]string{
		"fuser":   user,
		"fgroup":  group,
		"fprefix": key,
	}

	loginAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	request, err := NewMultipartRequest(endpoint, extraParams, file, loginAuth)
	FatalError("postFile", err)

	client := &http.Client{Timeout: 5}
	resp, err := client.Do(request)
	FatalError("postFile", err)

	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	FatalError("postFile", err)
	resp.Body.Close()

	PrintlnInfo("postFile:status", resp.StatusCode)
	PrintlnInfo("postFile:result", body)

	return nil
}

func clientDeleteFile(user, key, endpoint, auth string) error {
	if auth == "" {
		DebugInfo("clientDeleteFile", "auth is empty")
		return ErrUnauthorized
	}
	loginAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	deleteUrl := strings.Join([]string{endpoint, user, key}, "/")
	DebugInfo("deleteFile", deleteUrl)
	request, err := http.NewRequest("DELETE", deleteUrl, nil)
	request.Header.Add("Authorization", "Basic "+loginAuth)
	FatalError("clientDeleteFile", err)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(request)

	FatalError("deleteFile", err)
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	FatalError("deleteFile", err)
	resp.Body.Close()

	PrintlnInfo("deleteFile:status", resp.StatusCode)
	PrintlnInfo("deleteFile:result", body)

	return nil
}
