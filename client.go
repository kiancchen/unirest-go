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
	"strings"
	"unsafe"
)

type HTTPClient struct {
	query     url.Values
	form      url.Values
	files     []*fileField
	url       string
	path      string
	body      []byte
	method    string
	header    http.Header
	basicAuth [2]string
	makeCopy  bool
}

type fileField struct {
	key      string
	filename string
	content  []byte
}

func New() *HTTPClient {
	return &HTTPClient{
		query:    url.Values{},
		form:     url.Values{},
		method:   "GET",
		header:   http.Header{},
		files:    make([]*fileField, 0),
		makeCopy: true,
	}
}

func (c *HTTPClient) AutoClone(b bool) *HTTPClient {
	c = c.Clone()
	c.makeCopy = b
	return c
}

func (c *HTTPClient) SetURL(url string) *HTTPClient {
	if c.makeCopy {
		c = c.Clone()
	}
	c.url = strings.TrimRight(url, "/")
	return c
}

func (c *HTTPClient) AppendPath(path string) *HTTPClient {
	if c.makeCopy {
		c = c.Clone()
	}
	if path == "" {
		return c
	}
	if path[0] != '/' {
		path = "/" + path
	}
	path = strings.TrimRight(path, "/")
	c.path += path
	return c
}

func (c *HTTPClient) AddQuery(key, value string) *HTTPClient {
	if c.makeCopy {
		c = c.Clone()
	}
	c.query.Add(key, value)
	return c
}

func (c *HTTPClient) AddHeader(key, value string) *HTTPClient {
	if c.makeCopy {
		c = c.Clone()
	}
	c.header.Add(key, value)
	return c
}

func (c *HTTPClient) AddFormField(key, value string) *HTTPClient {
	if c.makeCopy {
		c = c.Clone()
	}
	c.form.Add(key, value)
	c.method = "POST"
	return c
}

func (c *HTTPClient) AddFile(key, filename string, content []byte) *HTTPClient {
	if c.makeCopy {
		c = c.Clone()
	}
	c.files = append(c.files, &fileField{
		key:      key,
		filename: filename,
		content:  content,
	})
	c.method = "POST"
	return c
}

func (c *HTTPClient) SetBasicAuth(username, password string) *HTTPClient {
	if c.makeCopy {
		c = c.Clone()
	}
	c.basicAuth[0] = username
	c.basicAuth[1] = password
	return c
}

func (c *HTTPClient) SetJSONBody(json []byte) *HTTPClient {
	if c.makeCopy {
		c = c.Clone()
	}
	c.body = json
	c.header.Set("Content-Type", "application/json")
	c.method = "POST"
	return c
}

func (c *HTTPClient) SetRawBody(body []byte) *HTTPClient {
	if c.makeCopy {
		c = c.Clone()
	}
	c.body = body
	c.method = "POST"
	c.header.Del("Content-Type")
	return c
}

func (c *HTTPClient) Get() *HTTPClient {
	if c.makeCopy {
		c = c.Clone()
	}
	c.method = "GET"
	return c
}

func (c *HTTPClient) Post() *HTTPClient {
	if c.makeCopy {
		c = c.Clone()
	}
	c.method = "POST"
	return c
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

	u, err := url.Parse(c.url + c.path)
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

func (c *HTTPClient) Clone() *HTTPClient {
	clone := *c
	clone.query = copyMap(c.query)
	clone.header = copyMap(c.header)
	clone.form = copyMap(c.form)
	return &clone
}

func copyMap(m map[string][]string) map[string][]string {
	clone := map[string][]string{}
	for key, values := range m {
		clone[key] = make([]string, len(values))
		copy(clone[key], values)
	}
	return clone
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
