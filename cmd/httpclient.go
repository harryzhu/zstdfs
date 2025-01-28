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

func clientDo(method, url, auth, fpath string, formKV map[string]string) error {
	if IsAnyEmpty(method, url) {
		return ErrParamEmpty
	}

	var err error
	var loginAuth string
	var request *http.Request

	if auth == "" {
		DebugInfo("clientDo", "auth is empty")
	} else {
		loginAuth = base64.StdEncoding.EncodeToString([]byte(auth))
	}

	if fpath == "" {
		request, err = http.NewRequest(method, url, nil)
	}

	if fpath != "" && strings.ToUpper(method) == "POST" {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)
		part, err := writer.CreateFormFile("file", filepath.Base(fpath))
		fp, err := os.Open(fpath)
		FatalError("clientDo", err)
		io.Copy(part, fp)

		for k, v := range formKV {
			writer.WriteField(k, v)
		}
		writer.Close()
		//
		request, err = http.NewRequest("POST", url, body)
		request.Header.Add("Content-Type", writer.FormDataContentType())
	}
	FatalError("clientDo", err)

	if loginAuth != "" {
		request.Header.Add("Authorization", "Basic "+loginAuth)
	}

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Do(request)

	FatalError("clientDo", err)
	body := &bytes.Buffer{}
	_, err = body.ReadFrom(resp.Body)
	FatalError("clientDo", err)
	resp.Body.Close()

	PrintlnInfo("clientDo:status", resp.StatusCode)
	PrintlnInfo("clientDo:result", body)

	return nil
}
