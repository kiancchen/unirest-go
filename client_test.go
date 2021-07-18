package unirest

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSimpleQuery(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := true
		query := r.URL.Query()
		v, ok := query["field1"]
		expected = expected && ok
		expected = expected && v[0] == "1"

		if expected {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	}))
	defer svr.Close()
	c, err := New().SetURL(svr.URL).AddQuery("field1", "1").Send().AsString()
	assert.NoError(t, err)
	assert.Equal(t, "true", c)
}

func TestMultiQuery(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := true
		query := r.URL.Query()
		v, ok := query["field1"]
		expected = expected && ok
		expected = expected && v[0] == "1"
		expected = expected && v[1] == "2"

		v, ok = query["field2"]
		expected = expected && ok
		expected = expected && v[0] == "1"

		if expected {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	}))
	defer svr.Close()
	c, err := New().SetURL(svr.URL).
		AddQuery("field1", "1").
		AddQuery("field1", "2").
		AddQuery("field2", "1").
		Send().
		AsString()

	assert.NoError(t, err)
	assert.Equal(t, "true", c)
}

func TestSimpleForm(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := true
		vv := r.PostFormValue("field2")
		expected = expected && vv == "1"

		v, ok := r.PostForm["field1"]
		expected = expected && ok
		expected = expected && v[0] == "1"
		expected = expected && v[1] == "2"

		if expected {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	}))
	defer svr.Close()
	c, err := New().SetURL(svr.URL).
		AddFormField("field1", "1").
		AddFormField("field1", "2").
		AddFormField("field2", "1").
		Send().
		AsString()

	assert.NoError(t, err)
	assert.Equal(t, "true", c)
}

func TestSimpleFile(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := true
		_, f, err := r.FormFile("file2")
		if err != nil {
			panic(err)
		}

		expected = expected && f.Filename == "file2.txt"
		file, _ := f.Open()
		b, _ := ioutil.ReadAll(file)
		expected = expected && string(b) == ("file2")

		fs := r.MultipartForm.File["file1"]
		file, _ = fs[0].Open()
		b, _ = ioutil.ReadAll(file)
		expected = expected && fs[0].Filename == "file1.txt"
		expected = expected && string(b) == ("file1")

		file, _ = fs[1].Open()
		b, _ = ioutil.ReadAll(file)
		expected = expected && fs[1].Filename == "file11.txt"
		expected = expected && string(b) == ("file11")

		if expected {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	}))
	defer svr.Close()
	c, err := New().SetURL(svr.URL).
		AddFile("file1", "file1.txt", []byte("file1")).
		AddFile("file1", "file11.txt", []byte("file11")).
		AddFile("file2", "file2.txt", []byte("file2")).
		Send().
		AsString()

	assert.NoError(t, err)
	assert.Equal(t, "true", c)
}

func TestSimpleBody(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := true
		b, _ := ioutil.ReadAll(r.Body)
		expected = expected && string(b) == "raw body"

		if expected {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	}))
	defer svr.Close()
	c, err := New().SetURL(svr.URL).SetRawBody([]byte("raw body")).Send().AsString()
	assert.NoError(t, err)
	assert.Equal(t, "true", c)
}

func TestSimpleJSON(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := true
		b, _ := ioutil.ReadAll(r.Body)
		expected = expected && string(b) == "{\"A\":1}"
		expected = expected && r.Header.Get("Content-Type") == "application/json"

		if expected {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	}))
	defer svr.Close()
	c, err := New().SetURL(svr.URL).SetJSONBody([]byte("{\"A\":1}")).Send().AsString()
	assert.NoError(t, err)
	assert.Equal(t, "true", c)
}

func TestBodyAndJSON(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		expected := true
		b, _ := ioutil.ReadAll(r.Body)
		expected = expected && string(b) == "raw body"
		expected = expected && r.Header.Get("Content-Type") == ""

		if expected {
			w.Write([]byte("true"))
		} else {
			w.Write([]byte("false"))
		}
	}))
	defer svr.Close()
	c, err := New().SetURL(svr.URL).SetJSONBody([]byte("{\"A\":1}")).SetRawBody([]byte("raw body")).Send().AsString()
	assert.NoError(t, err)
	assert.Equal(t, "true", c)
}

func TestClone(t *testing.T) {
	c := New()
	c2 := c.Post()
	assert.Equal(t, "GET", c.method)
	assert.Equal(t, "POST", c2.method)

	c3 := c.AddHeader("Q", "1")
	assert.Equal(t, 0, len(c.header["Q"]))
	assert.Equal(t, "1", c3.header["Q"][0])

	c4 := c.AddQuery("q", "1")
	assert.Equal(t, 0, len(c.query["q"]))
	assert.Equal(t, "1", c4.query["q"][0])

	c5 := c.SetRawBody([]byte("123"))
	assert.Equal(t, 0, len(c.body))
	assert.Equal(t, "123", string(c5.body))

	c6 := c.SetBasicAuth("1", "2")
	assert.Equal(t, "", c.basicAuth[0])
	assert.Equal(t, "", c.basicAuth[1])
	assert.Equal(t, "1", c6.basicAuth[0])
	assert.Equal(t, "2", c6.basicAuth[1])

	c7 := c.AddFile("1", "2", []byte("3"))
	assert.Equal(t, 0, len(c.files))
	assert.Equal(t, 1, len(c7.files))

	c = c.AutoClone(false)
	c8 := c.AddQuery("1", "2")
	assert.Equal(t, "2", c.query["1"][0])
	assert.Equal(t, "2", c8.query["1"][0])

	c8 = c.AutoClone(false).AddQuery("c8", "2")
	assert.Equal(t, 0, len(c.query["c8"]))
	assert.Equal(t, "2", c8.query["c8"][0])

	c8 = c.AutoClone(true).AddQuery("c88", "2")
	assert.Equal(t, 0, len(c.query["c88"]))
	assert.Equal(t, "2", c8.query["c88"][0])

	return
}

func TestAddPath(t *testing.T) {
	c := New().SetURL("https://a.com/").
		AppendPath("p1").
		AppendPath("/p2").
		AppendPath("").
		AppendPath("/p3/").
		AppendPath("/p4/")
	assert.Equal(t, "https://a.com/p1/p2/p3/p4", c.url+c.path)
}

func TestExpectedError(t *testing.T) {
	_, err := New().AddFormField("1", "1").SetRawBody([]byte("123")).Send().AsBytes()
	assert.Error(t, err, "unirest-go: can send this request with multiple content type")
}
