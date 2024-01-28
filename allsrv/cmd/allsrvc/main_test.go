package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http/httptest"
	"testing"

	"github.com/jsteenb2/mess/allsrv"
	"github.com/jsteenb2/mess/allsrv/allsrvtesting"
)

func TestCliSVC(t *testing.T) {
	allsrvtesting.TestSVC(t, func(t *testing.T, opts allsrvtesting.SVCTestOpts) allsrvtesting.SVCDeps {
		svc := allsrvtesting.NewInmemSVC(t, opts)
		srv := httptest.NewServer(allsrv.NewServerV2(svc))
		t.Cleanup(srv.Close)

		return allsrvtesting.SVCDeps{
			SVC: &cmdCLI{addr: srv.URL},
		}
	})
}

type cmdCLI struct {
	addr string
}

func (c *cmdCLI) CreateFoo(ctx context.Context, f allsrv.Foo) (allsrv.Foo, error) {
	return c.expectFoo(ctx, "add", "--name", f.Name, "--note", f.Note)
}

func (c *cmdCLI) ReadFoo(ctx context.Context, id string) (allsrv.Foo, error) {
	return c.expectFoo(ctx, "read", id)
}

func (c *cmdCLI) UpdateFoo(ctx context.Context, f allsrv.FooUpd) (allsrv.Foo, error) {
	args := []string{"--id", f.ID}
	if f.Name != nil {
		args = append(args, "--name", *f.Name)
	}
	if f.Note != nil {
		args = append(args, "--note", *f.Note)
	}
	return c.expectFoo(ctx, "update", args...)
}

func (c *cmdCLI) DelFoo(ctx context.Context, id string) error {
	_, err := c.execute(ctx, "rm", id)
	return err
}

func (c *cmdCLI) expectFoo(ctx context.Context, op string, args ...string) (allsrv.Foo, error) {
	b, err := c.execute(ctx, op, args...)
	if err != nil {
		return allsrv.Foo{}, err
	}

	var out allsrv.Foo
	if err := json.Unmarshal(b, &out); err != nil {
		return allsrv.Foo{}, err
	}

	return out, nil
}

func (c *cmdCLI) execute(ctx context.Context, op string, args ...string) ([]byte, error) {
	cmd := newCmd()

	var buf bytes.Buffer
	cmd.SetOut(&buf)
	cmd.SetErr(&buf)

	cmd.SetArgs(append([]string{op, "--addr", c.addr}, args...))

	err := cmd.ExecuteContext(ctx)
	return buf.Bytes(), err
}
