package httpclient

import (
	"bytes"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"net/url"
	"os"
	"strings"
)

func BuildRequestUploadFiles(uri string, queryParams, postParams map[string]string, files map[string][]string) (*http.Request, error) {
	return BuildRequest(http.MethodPost, uri, queryParams, postParams, files, nil, nil)
}

func BuildRequestUploadFilesFromIOReader(uri string, queryParams, postParams map[string]string, files map[string]map[string]io.Reader) (*http.Request, error) {
	return BuildRequest(http.MethodPost, uri, queryParams, postParams, nil, files, nil)
}

func BuildPostRequest(uri string, queryParams, postParams map[string]string) (*http.Request, error) {
	return BuildRequest(http.MethodPost, uri, queryParams, postParams, nil, nil, nil)
}

func BuildGetRequest(uri string, queryParams map[string]string) (*http.Request, error) {
	return BuildRequest(http.MethodGet, uri, queryParams, nil, nil, nil, nil)
}

func BuildJSONRequest(uri string, queryParams map[string]string, jsonBody io.Reader) (*http.Request, error) {
	return BuildRequest(http.MethodPost, uri, queryParams, nil, nil, nil, jsonBody)
}

func BuildRequest(method, uri string, queryParams, postParams map[string]string, files map[string][]string, ioReaders map[string]map[string]io.Reader, jsonBody io.Reader) (req *http.Request, err error) {
	switch {
	case files != nil || ioReaders != nil:
		body, contentType, err := buildMultiplerPart(postParams, files, ioReaders)
		if err != nil {
			return nil, err
		}
		req, err = http.NewRequest(method, uri, body)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", contentType)

	case jsonBody != nil:
		req, err = http.NewRequest(method, uri, jsonBody)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/json")
	case files == nil && postParams != nil:
		form := url.Values{}
		for k, v := range postParams {
			form.Add(k, v)
		}
		req, err = http.NewRequest(method, uri, strings.NewReader(form.Encode()))
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	default:
		req, err = http.NewRequest(method, uri, nil)
		if err != nil {
			return nil, err
		}
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	}
	q := req.URL.Query()
	for k, v := range queryParams {
		q.Add(k, v)
	}
	req.URL.RawQuery = q.Encode()
	return req, nil
}

func buildMultiplerPart(params map[string]string, files map[string][]string, ioReaders map[string]map[string]io.Reader) (*bytes.Buffer, string, error) {
	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	for name, ps := range files {
		for _, p := range ps {
			if err := writeFileToWriter(writer, name, p); err != nil {
				return nil, "", err
			}
		}
	}

	for fieldName, readers := range ioReaders {
		for filename, r := range readers {
			if err := writeReaderToWriter(writer, fieldName, filename, r); err != nil {
				return nil, "", err
			}
		}
	}
	for key, val := range params {
		if err := writer.WriteField(key, val); err != nil {
			return nil, "", err
		}
	}
	if err := writer.Close(); err != nil {
		return nil, "", err
	}
	return body, writer.FormDataContentType(), nil
}

func writeFileToWriter(writer *multipart.Writer, name, fpath string) error {
	f, err := os.Open(fpath)
	if err != nil {
		return err
	}
	defer f.Close()
	fs, err := f.Stat()
	if err != nil {
		return err
	}

	header := textproto.MIMEHeader{}
	buffer := make([]byte, 512)
	_, err = f.Read(buffer)
	if err != nil {
		return err
	}
	f.Seek(0, 0)
	header.Add("Content-Type", http.DetectContentType(buffer))
	header.Add("Content-Disposition", fmt.Sprintf(`form-data; filename="%s"; name="%s"`, fs.Name(), name))

	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, f)
	if err != nil {
		return err
	}
	return nil
}

func writeReaderToWriter(writer *multipart.Writer, filedName, filename string, reader io.Reader) error {
	header := textproto.MIMEHeader{}
	first512 := make([]byte, 512)
	_, err := reader.Read(first512)
	if err != nil {
		return err
	}
	buffer := bytes.NewBuffer(first512)
	header.Add("Content-Type", http.DetectContentType(first512))
	header.Add("Content-Disposition", fmt.Sprintf(`form-data; filename="%s"; name="%s"`, filename, filedName))

	part, err := writer.CreatePart(header)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, buffer)
	if err != nil {
		return err
	}
	_, err = io.Copy(part, reader)
	if err != nil {
		return err
	}
	return nil
}
