package httpclient_test

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"

	. "github.com/nullne/utils/httpclient"
)

var (
	gQueryParams map[string]string   = map[string]string{"query-a": "a", "query-b": "b"}
	gPostParams  map[string]string   = map[string]string{"post-a": "a", "post-b": "b"}
	gFiles       map[string][]string = map[string][]string{"two-file": []string{"./testdata/sample.jpg", "./testdata/sample.txt"}, "single-file": []string{"./testdata/sample.webp"}}
)

var testClient *http.Client = &http.Client{Transport: &http.Transport{}}

func ExampleF_BuildRequestUploadFiles() {
	req, err := BuildRequestUploadFiles(localURL, gQueryParams, gPostParams, gFiles)
	handleErrorAndRequest(req, err)
	// Output:
	// URL query:query-a=>[a],query-b=>[b],
	// URL values:query-a=>[a],query-b=>[b],
	// Multipart Form values:post-a=>[a],post-b=>[b],
	// Multipart Form files:single-file=>[sample.webp: Content-Disposition=>[form-data; filename="sample.webp"; name="single-file"],Content-Type=>[image/webp],],two-file=>[sample.jpg: Content-Disposition=>[form-data; filename="sample.jpg"; name="two-file"],Content-Type=>[image/jpeg], sample.txt: Content-Disposition=>[form-data; filename="sample.txt"; name="two-file"],Content-Type=>[application/octet-stream],],
}

func ExampleF_BuildRequestUploadFilesFromIOReader() {
	reader, err := os.Open("./testdata/sample.webp")
	if err != nil {
		log.Fatal(err)
	}
	files := make(map[string]map[string]io.Reader)
	files["field1"] = map[string]io.Reader{"file1": reader}
	req, err := BuildRequestUploadFilesFromIOReader(localURL, gQueryParams, gPostParams, files)
	handleErrorAndRequest(req, err)
	// Output:
	// URL query:query-a=>[a],query-b=>[b],
	// URL values:query-a=>[a],query-b=>[b],
	// Multipart Form values:post-a=>[a],post-b=>[b],
	// Multipart Form files:field1=>[file1: Content-Disposition=>[form-data; filename="file1"; name="field1"],Content-Type=>[image/webp],],
}

func ExampleF_BuildPostRequest() {
	req, err := BuildPostRequest(localURL, gQueryParams, gPostParams)
	handleErrorAndRequest(req, err)
	// Output:
	// URL query:query-a=>[a],query-b=>[b],
	// URL values:post-a=>[a],post-b=>[b],query-a=>[a],query-b=>[b],
}

func ExampleF_BuildGetRequest() {
	req, err := BuildGetRequest(localURL, gQueryParams)
	handleErrorAndRequest(req, err)
	// Output:
	// URL query:query-a=>[a],query-b=>[b],
	// URL values:query-a=>[a],query-b=>[b],
}

func ExampleF_BuildJSONRequest() {
	jb := bytes.NewBuffer([]byte(`{color:"red",value:"#f00"}`))
	req, err := BuildJSONRequest(localURL, gQueryParams, jb)
	handleErrorAndRequest(req, err)
	// Output:
	// URL query:query-a=>[a],query-b=>[b],
	// JSON: {color:"red",value:"#f00"}
	// URL values:query-a=>[a],query-b=>[b],
}

func handleErrorAndRequest(req *http.Request, err error) {
	if err != nil {
		log.Fatal(err)
	}
	resp, err := testClient.Do(req)
	if err != nil {
		fmt.Println(err)
		return
	} else {
		data, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			log.Fatal(err)
		}
		resp.Body.Close()
		fmt.Println(string(data))
	}
}
