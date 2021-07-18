package unirest

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"mime/multipart"
	"net/http"
	"net/url"
	"reflect"
	"unsafe"
)

func copyMap(m map[string][]string) map[string][]string {
	clone := map[string][]string{}
	for key, values := range m {
		clone[key] = make([]string, len(values))
		copy(clone[key], values)
	}
	return clone
}

type HTTPClient struct {
	query     url.Values
	form      url.Values
	files     []*fileField
	url       string
	body      []byte
	method    string
	header    http.Header
	basicAuth [2]string
}

type fileField struct {
	key      string
	filename string
	content  []byte
}

func New(rawUrl string) *HTTPClient {
	return &HTTPClient{
		url:    rawUrl,
		query:  url.Values{},
		form:   url.Values{},
		method: "GET",
		header: http.Header{},
		files:  make([]*fileField, 0),
	}
}

func (c *HTTPClient) clone() *HTTPClient {
	clone := *c
	clone.query = copyMap(c.query)
	clone.header = copyMap(c.header)
	clone.form = copyMap(c.form)
	return &clone
}

func (c *HTTPClient) AddQuery(key, value string) *HTTPClient {
	clone := c.clone()
	clone.query.Add(key, value)
	return clone
}

func (c *HTTPClient) AddHeader(key, value string) *HTTPClient {
	clone := c.clone()
	clone.header.Add(key, value)
	return clone
}

func (c *HTTPClient) AddFormField(key, value string) *HTTPClient {
	clone := c.clone()
	clone.form.Add(key, value)
	clone.method = "POST"
	return clone
}

func (c *HTTPClient) AddFile(key, filename string, content []byte) *HTTPClient {
	clone := c.clone()
	clone.files = append(c.files, &fileField{
		key:      key,
		filename: filename,
		content:  content,
	})
	clone.method = "POST"
	return clone
}

func (c *HTTPClient) SetBasicAuth(username, password string) *HTTPClient {
	clone := c.clone()
	clone.basicAuth[0] = username
	clone.basicAuth[1] = password
	return clone
}

func (c *HTTPClient) SetJSONBody(json []byte) *HTTPClient {
	clone := c.clone()
	clone.body = json
	clone.header.Set("Content-Type", "application/json")
	clone.method = "POST"
	return clone
}

func (c *HTTPClient) SetRawBody(body []byte) *HTTPClient {
	clone := c.clone()
	clone.body = body
	clone.method = "POST"
	clone.header.Del("Content-Type")
	return clone
}

func (c *HTTPClient) Get() *HTTPClient {
	clone := c.clone()
	clone.method = "GET"
	return clone
}

func (c *HTTPClient) Post() *HTTPClient {
	clone := c.clone()
	clone.method = "POST"
	return clone
}

func (c *HTTPClient) Send() *Response {
	req, err := c.ParseRequest()
	if err != nil {
		return &Response{Err: err}
	}

	var httpclient http.Client
	resp, err := httpclient.Do(req)
	if err != nil {
		return &Response{Err: err}
	}

	return &Response{Response: resp}
}

func (c *HTTPClient) ParseRequest() (*http.Request, error) {
	req := &http.Request{
		Method: c.method,
		Header: c.header,
	}

	req.Header.Set("User-Agent", "Unirest-Go/1.0")

	if c.basicAuth[0] != "" {
		req.SetBasicAuth(c.basicAuth[0], c.basicAuth[1])
	}

	u, err := url.Parse(c.url)
	if err != nil {
		return nil, err
	}
	req.URL = u

	if len(c.query) != 0 {
		req.URL.RawQuery = c.query.Encode()
	}

	if c.body != nil && (len(c.files) != 0 || len(c.form) != 0) {
		return nil, errors.New("unirest-go: can send this request with multiple content type")
	}

	var reader *bytes.Reader
	if c.body != nil {
		reader = bytes.NewReader(c.body)
	} else if len(c.files) != 0 {
		body := &bytes.Buffer{}
		writer := multipart.NewWriter(body)

		for _, file := range c.files {
			fw, err := writer.CreateFormFile(file.key, file.filename)
			if err != nil {
				return nil, err
			}
			_, err = fw.Write(file.content)
			if err != nil {
				return nil, err
			}
		}

		for key, values := range c.form {
			for _, value := range values {
				err := writer.WriteField(key, value)
				if err != nil {
					return nil, err
				}
			}
		}

		err := writer.Close()
		if err != nil {
			return nil, err
		}

		reader = bytes.NewReader(body.Bytes())
		req.Header.Set("Content-Type", writer.FormDataContentType())
	} else if len(c.form) != 0 {
		reader = bytes.NewReader(s2b(c.form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	if reader != nil {
		readCloser := ioutil.NopCloser(reader)
		req.Body = readCloser
		req.ContentLength = int64(reader.Len())
		snapshot := *reader
		req.GetBody = func() (io.ReadCloser, error) {
			r := snapshot
			return ioutil.NopCloser(&r), nil
		}
	}
	return req, nil
}

// b2s converts byte slice to a string without memory allocation.
// See https://groups.google.com/forum/#!msg/Golang-Nuts/ENgbUzYvCuU/90yGx7GUAgAJ .
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func b2s(b []byte) string {
	/* #nosec G103 */
	return *(*string)(unsafe.Pointer(&b))
}

// s2b converts string to a byte slice without memory allocation.
//
// Note it may break if string and/or slice header will change
// in the future go versions.
func s2b(s string) (b []byte) {
	/* #nosec G103 */
	bh := (*reflect.SliceHeader)(unsafe.Pointer(&b))
	/* #nosec G103 */
	sh := *(*reflect.StringHeader)(unsafe.Pointer(&s))
	bh.Data = sh.Data
	bh.Len = sh.Len
	bh.Cap = sh.Len
	return b
}
