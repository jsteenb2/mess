package allsrv

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
	
	"github.com/jsteenb2/errors"
)

type ClientHTTP struct {
	addr string
	c    *http.Client
}

var _ SVC = (*ClientHTTP)(nil)

func NewClientHTTP(addr string, c *http.Client) *ClientHTTP {
	return &ClientHTTP{
		addr: addr,
		c:    c,
	}
}

func (c *ClientHTTP) CreateFoo(ctx context.Context, f Foo) (Foo, error) {
	req, err := jsonReq(ctx, "POST", c.fooPath(""), toReqCreateFooV1(f))
	if err != nil {
		return Foo{}, InternalErr(err.Error())
	}
	return returnsFooReq(c.c, req)
}

func (c *ClientHTTP) ReadFoo(ctx context.Context, id string) (Foo, error) {
	if id == "" {
		return Foo{}, errIDRequired
	}

	req, err := http.NewRequestWithContext(ctx, "GET", c.fooPath(id), nil)
	if err != nil {
		return Foo{}, InternalErr(err.Error())
	}
	return returnsFooReq(c.c, req)
}

func (c *ClientHTTP) UpdateFoo(ctx context.Context, f FooUpd) (Foo, error) {
	req, err := jsonReq(ctx, "PATCH", c.fooPath(f.ID), toReqUpdateFooV1(f))
	if err != nil {
		return Foo{}, InternalErr(err.Error())
	}
	return returnsFooReq(c.c, req)
}

func (c *ClientHTTP) DelFoo(ctx context.Context, id string) error {
	if id == "" {
		return errIDRequired
	}

	req, err := http.NewRequestWithContext(ctx, "DELETE", c.fooPath(id), nil)
	if err != nil {
		return InternalErr(err.Error())
	}

	_, err = doReq[any](c.c, req)
	return err
}

func (c *ClientHTTP) fooPath(id string) string {
	u := c.addr + "/v1/foos"
	if id == "" {
		return u
	}
	return u + "/" + id
}

func jsonReq(ctx context.Context, method, path string, v any) (*http.Request, error) {
	var buf bytes.Buffer
	if err := json.NewEncoder(&buf).Encode(v); err != nil {
		return nil, InvalidErr("failed to marshal payload: " + err.Error())
	}

	req, err := http.NewRequestWithContext(ctx, method, path, &buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	return req, nil
}

func returnsFooReq(c *http.Client, req *http.Request) (Foo, error) {
	data, err := doReq[ResourceFooAttrs](c, req)
	if err != nil {
		return Foo{}, err
	}
	return toFoo(data), nil
}

func doReq[Attr Attrs](c *http.Client, req *http.Request) (Data[Attr], error) {
	resp, err := c.Do(req)
	if err != nil {
		return *new(Data[Attr]), InternalErr(err.Error())
	}
	defer func() {
		io.Copy(io.Discard, resp.Body)
		resp.Body.Close()
	}()

	if resp.Header.Get("Content-Type") != "application/json" {
		b, err := io.ReadAll(io.LimitReader(resp.Body, 500<<10))
		if err != nil {
			return *new(Data[Attr]), InternalErr("failed to read response body: ", err.Error())
		}
		return *new(Data[Attr]), InternalErr("invalid content type received; content=" + string(b))
	}
	// TODO(berg): handle unexpected status code (502|503|etc)

	var respBody RespBody[Attr]
	err = json.NewDecoder(resp.Body).Decode(&respBody)
	if err != nil {
		return *new(Data[Attr]), InternalErr(err.Error())
	}

	var errs []error
	for _, respErr := range respBody.Errs {
		errs = append(errs, toErr(respErr))
	}
	if len(errs) == 1 {
		return *new(Data[Attr]), errs[0]
	}
	if len(errs) > 1 {
		return *new(Data[Attr]), errors.Join(errs)
	}

	if respBody.Data == nil {
		return *new(Data[Attr]), nil
	}

	return *respBody.Data, nil
}

func toReqCreateFooV1(f Foo) ReqCreateFooV1 {
	return ReqCreateFooV1{
		Data: Data[FooCreateAttrs]{
			Type: "foo",
			Attrs: FooCreateAttrs{
				Name: f.Name,
				Note: f.Note,
			},
		},
	}
}

func toReqUpdateFooV1(f FooUpd) ReqUpdateFooV1 {
	return ReqUpdateFooV1{
		Data: Data[FooUpdAttrs]{
			Type: "foo",
			ID:   f.ID,
			Attrs: FooUpdAttrs{
				Name: f.Name,
				Note: f.Note,
			},
		},
	}
}

func toFoo(d Data[ResourceFooAttrs]) Foo {
	return Foo{
		ID:        d.ID,
		Name:      d.Attrs.Name,
		Note:      d.Attrs.Note,
		CreatedAt: toTime(d.Attrs.CreatedAt),
		UpdatedAt: toTime(d.Attrs.UpdatedAt),
	}
}

func toErr(respErr RespErr) error {
	errFn := InternalErr
	switch respErr.Code {
	case errCodeExist:
		errFn = ExistsErr
	case errCodeInvalid:
		errFn = InvalidErr
	case errCodeNotFound:
		errFn = NotFoundErr
	case errCodeUnAuthed:
		errFn = unauthedErr
	}
	var fields []any
	if respErr.Source != nil {
		fields = append(fields, "err_source", *respErr.Source)
	}
	return errFn(respErr.Msg, fields...)
}

func toTime(in string) time.Time {
	t, _ := time.Parse(time.RFC3339, in)
	return t
}
