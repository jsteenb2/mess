package main

import (
	"encoding/json"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

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

type cli struct {
	addr string
	id   string
	name string
	note string
}

func (c *cli) cmd() *cobra.Command {
	cmd := cobra.Command{
		Use: "allsrvc",
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
	cmd.Flags().StringVar(&c.addr, "addr", "http://localhost:8091", "addr for foo svc")
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

			return json.NewEncoder(cmd.OutOrStderr()).Encode(f)
		},
	}
	cmd.Flags().StringVar(&c.addr, "addr", "http://localhost:8091", "addr for foo svc")

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

			return json.NewEncoder(cmd.OutOrStderr()).Encode(f)
		},
	}
	cmd.Flags().StringVar(&c.addr, "addr", "http://localhost:8091", "addr for foo svc")
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
	cmd.Flags().StringVar(&c.addr, "addr", "http://localhost:8091", "addr for foo svc")

	return &cmd
}

func (c *cli) newClient() *allsrv.ClientHTTP {
	return allsrv.NewClientHTTP(c.addr, &http.Client{Timeout: 5 * time.Second})
}
