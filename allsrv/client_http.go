package allsrv

import (
	"context"
	"net/http"
	"time"

	"github.com/jsteenb2/errors"

	"github.com/jsteenb2/allsrvc"
)

type ClientHTTP struct {
	c *allsrvc.ClientHTTP
}

var _ SVC = (*ClientHTTP)(nil)

func NewClientHTTP(addr, origin string, c *http.Client, opts ...func(*allsrvc.ClientHTTP)) *ClientHTTP {
	return &ClientHTTP{
		c: allsrvc.NewClientHTTP(addr, origin, c, opts...),
	}
}

func (c *ClientHTTP) CreateFoo(ctx context.Context, f Foo) (Foo, error) {
	resp, err := c.c.CreateFoo(ctx, allsrvc.FooCreateAttrs{
		Name: f.Name,
		Note: f.Note,
	})
	if err != nil {
		return Foo{}, InternalErr(err.Error())
	}
	newFoo, err := takeRespFoo(resp)
	return newFoo, errors.Wrap(err)
}

func (c *ClientHTTP) ReadFoo(ctx context.Context, id string) (Foo, error) {
	resp, err := c.c.ReadFoo(ctx, id)
	if err != nil {
		if errors.Is(err, allsrvc.ErrIDRequired) {
			return Foo{}, errIDRequired
		}
	}

	newFoo, err := takeRespFoo(resp)
	return newFoo, errors.Wrap(err)
}

func (c *ClientHTTP) UpdateFoo(ctx context.Context, f FooUpd) (Foo, error) {
	resp, err := c.c.UpdateFoo(ctx, f.ID, allsrvc.FooUpdAttrs{
		Name: f.Name,
		Note: f.Note,
	})
	if err != nil {
		return Foo{}, InternalErr(err.Error())
	}
	newFoo, err := takeRespFoo(resp)
	return newFoo, errors.Wrap(err)
}

func (c *ClientHTTP) DelFoo(ctx context.Context, id string) error {
	resp, err := c.c.DelFoo(ctx, id)
	if err != nil {
		if errors.Is(err, allsrvc.ErrIDRequired) {
			return errIDRequired
		}
	}

	return errors.Wrap(convertSDKErrors(resp.Errs))
}

func DataToFoo(data allsrvc.Data[allsrvc.ResourceFooAttrs]) Foo {
	return Foo{
		ID:        data.ID,
		Name:      data.Attrs.Name,
		Note:      data.Attrs.Note,
		CreatedAt: toTime(data.Attrs.CreatedAt),
		UpdatedAt: toTime(data.Attrs.UpdatedAt),
	}
}

func takeRespFoo(respBody allsrvc.RespBody[allsrvc.ResourceFooAttrs]) (Foo, error) {
	if err := convertSDKErrors(respBody.Errs); err != nil {
		return Foo{}, errors.Wrap(err)
	}

	if respBody.Data == nil {
		return Foo{}, nil
	}

	return DataToFoo(*respBody.Data), nil
}

func convertSDKErrors(errs []allsrvc.RespErr) error {
	// TODO(@berg): update this to slices pkg when 1.23 lands
	switch out := toSlc(errs, toErr); {
	case len(out) == 1:
		return out[0]
	case len(out) > 1:
		return errors.Join(out)
	default:
		return nil
	}
}

func toErr(respErr allsrvc.RespErr) error {
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

func toSlc[In, Out any](in []In, to func(In) Out) []Out {
	out := make([]Out, len(in))
	for _, v := range in {
		out = append(out, to(v))
	}
	return out
}
