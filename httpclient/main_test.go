package httpclient_test

import (
	"bytes"
	"flag"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"net/url"
	"os"
	"sort"
	"strings"
	"testing"
)

var localURL string

func printMapInOrder(it interface{}) string {
	maps := make(map[string]interface{})
	switch i := it.(type) {
	case map[string][]*multipart.FileHeader:
		for k, v := range i {
			fhs := make([]string, 0, len(v))
			for _, fh := range v {
				fhs = append(fhs, fmt.Sprintf("%s: %s", fh.Filename, printMapInOrder(fh.Header)))
			}
			maps[k] = fhs
		}
	case url.Values:
		for k, v := range i {
			maps[k] = v
		}
	case textproto.MIMEHeader:
		for k, v := range i {
			maps[k] = v
		}
	case map[string][]string:
		for k, v := range i {
			maps[k] = v
		}
	default:
		return "unsupported type"
	}
	var buff bytes.Buffer
	keys := make([]string, 0, len(maps))
	for k, _ := range maps {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&buff, "%s=>%v,", k, maps[k])
	}
	return buff.String()
}

func startServer() *httptest.Server {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		// fmt.Fprintln(w, "Content Type: ", req.Header.Get("Content-Type"))
		fmt.Fprintf(w, "URL query:%s\n", printMapInOrder(req.URL.Query()))
		if err := req.ParseForm(); err != nil {
			fmt.Fprintln(w, "Failed to parse form: ", err.Error())
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}

		if req.Header.Get("Content-Type") == "application/json" {
			data, err := ioutil.ReadAll(req.Body)
			if err != nil {
				fmt.Fprintln(w, "Failed read from json body: ", err.Error())
				w.WriteHeader(http.StatusUnprocessableEntity)
				return
			}
			fmt.Fprintln(w, "JSON:", string(data))
		}
		fmt.Fprintf(w, "URL values:%s\n", printMapInOrder(req.Form))
		if !strings.Contains(req.Header.Get("Content-Type"), "multipart/form-data") {
			return
		}
		err := req.ParseMultipartForm(100 * 1024 * 1024)
		if err != nil {
			fmt.Fprint(w, "Failed to parse mutipartForm: ", err.Error())
			w.WriteHeader(http.StatusUnprocessableEntity)
			return
		}
		if req.MultipartForm == nil {
			return
		}
		fmt.Fprint(w, "Multipart Form values:")
		fmt.Fprintf(w, "%s\n", printMapInOrder(req.MultipartForm.Value))
		fmt.Fprint(w, "Multipart Form files:")
		fmt.Fprintf(w, "%s\n", printMapInOrder(req.MultipartForm.File))
	}))
	localURL = ts.URL
	return ts
}

func TestMain(m *testing.M) {
	flag.Parse()
	ts := startServer()
	defer ts.Close()
	os.Exit(m.Run())
}
