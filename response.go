package unirest

import (
	"errors"
	"io/ioutil"
	"net/http"
)

type Response struct {
	*http.Response
	Err error
}

func (r *Response) AsBytes() ([]byte, error) {
	if r.Err != nil {
		return nil, r.Err
	}

	if r.Response == nil {
		return nil, errors.New("the request is not sent")
	}

	buf, err := ioutil.ReadAll(r.Response.Body)
	r.Response.Body.Close()
	if err != nil {
		return nil, err
	}
	return buf, nil
}

func (r *Response) AsString() (string, error) {
	buf, err := r.AsBytes()
	if err != nil {
		return "", err
	}

	return b2s(buf), nil
}
