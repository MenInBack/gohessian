package gohessian

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"reflect"
)

type hessianRequest struct {
	body []byte
}

// NewClient return a client for hessian
func NewClient(host, url string) (c *Client) {
	host = HostCheck(host)
	return &Client{
		Host: host,
		URL:  url,
	}
}

// String format hessian client request location
func (c Client) String() string {
	return c.Host + c.URL
}

// Invoke send a request to hessian service and return the result of response
func (c *Client) Invoke(method string, params ...interface{}) (interface{}, error) {
	reqURL := c.Host + c.URL
	r := &hessianRequest{}
	r.packHead(method)
	for _, v := range params {
		r.packParam(v)
	}
	r.packEnd()

	resp, err := httpPost(reqURL, bytes.NewReader(r.body))
	if err != nil {
		fmt.Println("got hessian service response failed:", err)
		return nil, err
	}
	fmt.Println("got hessian service response success")

	if len(resp) == 0 {
		return nil, errors.New("method or params error, resp is null")
	}

	h := NewHessian(bytes.NewReader(resp))
	v, err := h.Parse()

	if err != nil {
		fmt.Println("hessian parse error", err)
		return nil, err
	}

	c.replyMap = v
	return v, nil
}

// BindResult bind reply to v, v must be a pointer
func (c *Client) BindResult(v interface{}) error {
	if reflect.ValueOf(v).Kind() != reflect.Ptr {
		return errors.New("not a pointer")
	}
	if reflect.ValueOf(v).IsNil() {
		return errors.New("nil pointer")
	}

	c.replyData = extractData(reflect.ValueOf(c.replyMap), reflect.TypeOf(v))
	reflect.ValueOf(v).Elem().Set(c.replyData.Elem())
	return nil
}

//httpPost send HTTP POST request, return bytes in body
func httpPost(url string, body io.Reader) (rb []byte, err error) {
	var resp *http.Response
	if resp, err = http.Post(url, "application/binary", body); err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		err = fmt.Errorf(resp.Status)
		return
	}
	defer resp.Body.Close()
	rb, err = ioutil.ReadAll(resp.Body)
	return
}

// packHead pack hessian request head
func (h *hessianRequest) packHead(method string) {
	tmp_b, _ := PackUint16(uint16(len(method)))
	h.body = append(h.body, []byte{99, 0, 1, 109}...)
	h.body = append(h.body, tmp_b...)
	h.body = append(h.body, []byte(method)...)
}

// packParam pack param in hessian request
func (h *hessianRequest) packParam(p Any) {
	tmp_b, err := Encode(p)
	if err != nil {
		panic(err)
	}
	h.body = append(h.body, tmp_b...)
}

// packEnd pack end of hessian request
func (h *hessianRequest) packEnd() {
	h.body = append(h.body, 'z')
}
