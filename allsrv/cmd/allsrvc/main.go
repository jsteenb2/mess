package main

import (
	"encoding/json"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/jsteenb2/errors"
	"github.com/spf13/cobra"

	"github.com/jsteenb2/allsrvc"
	"github.com/jsteenb2/mess/allsrv"
)

func main() {
	cmd := newCmd()
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}

func newCmd() *cobra.Command {
	c := new(cli)
	return c.cmd()
}

const name = "allsrvc"

type cli struct {
	// base flags
	addr string
	pass string
	user string

	// foo flags
	id   string
	name string
	note string
}

func (c *cli) cmd() *cobra.Command {
	cmd := cobra.Command{
		Use:          name,
		SilenceUsage: true,
	}

	cmd.AddCommand(
		c.cmdCreateFoo(),
		c.cmdReadFoo(),
		c.cmdUpdateFoo(),
		c.cmdRmFoo(),
	)

	return &cmd
}

func (c *cli) cmdCreateFoo() *cobra.Command {
	cmd := cobra.Command{
		Use:     "add",
		Aliases: []string{"create"},
		Short:   "creates a new foo",
		Args:    cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := c.newClient()

			f, err := client.CreateFoo(cmd.Context(), allsrv.Foo{
				Name: c.name,
				Note: c.note,
			})
			if err != nil {
				return err
			}

			return json.NewEncoder(cmd.OutOrStderr()).Encode(f)
		},
	}
	c.registerCommonFlags(&cmd)
	cmd.Flags().StringVar(&c.name, "name", "", "name of the new foo")
	cmd.Flags().StringVar(&c.note, "note", "", "optional foo note")

	return &cmd
}

func (c *cli) cmdReadFoo() *cobra.Command {
	cmd := cobra.Command{
		Use:   "read $FOO_ID",
		Short: "read a food by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := c.newClient()

			f, err := client.ReadFoo(cmd.Context(), args[0])
			if err != nil {
				return err
			}

			return errors.Wrap(writeFoo(cmd.OutOrStdout(), f))
		},
	}
	c.registerCommonFlags(&cmd)
	return &cmd
}

func (c *cli) cmdUpdateFoo() *cobra.Command {
	cmd := cobra.Command{
		Use:   "update",
		Short: "updates an existing foo",
		Args:  cobra.NoArgs,
		RunE: func(cmd *cobra.Command, args []string) error {
			client := c.newClient()

			upd := allsrv.FooUpd{
				ID: c.id,
			}
			if c.name != "" {
				upd.Name = &c.name
			}
			if c.note != "" {
				upd.Note = &c.note
			}

			f, err := client.UpdateFoo(cmd.Context(), upd)
			if err != nil {
				return err
			}

			return errors.Wrap(writeFoo(cmd.OutOrStdout(), f))
		},
	}
	c.registerCommonFlags(&cmd)
	cmd.Flags().StringVar(&c.id, "id", "", "id of the foo resource")
	cmd.Flags().StringVar(&c.name, "name", "", "optional foo name")
	cmd.Flags().StringVar(&c.note, "note", "", "optional foo note")

	return &cmd
}

func (c *cli) cmdRmFoo() *cobra.Command {
	cmd := cobra.Command{
		Use:   "rm $FOO_ID",
		Short: "delete a foo by id",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			client := c.newClient()
			return client.DelFoo(cmd.Context(), args[0])
		},
	}
	c.registerCommonFlags(&cmd)
	return &cmd
}

func (c *cli) newClient() *allsrv.ClientHTTP {
	return allsrv.NewClientHTTP(
		c.addr,
		name,
		&http.Client{Timeout: 5 * time.Second},
		allsrvc.WithBasicAuth(c.user, c.pass),
	)
}

func (c *cli) registerCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringVar(&c.addr, "addr", "http://localhost:8091", "addr for foo svc")
	cmd.Flags().StringVar(&c.user, "user", "admin", "user for basic auth")
	cmd.Flags().StringVar(&c.pass, "password", "pass", "password for basic auth")
}

func writeFoo(w io.Writer, f allsrv.Foo) error {
	err := json.NewEncoder(w).Encode(allsrv.FooToData(f))
	return errors.Wrap(err)
}
