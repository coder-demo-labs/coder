package cli

import (
	"fmt"

	"github.com/coder/serpent"
)

// yes prints "yes" and exits, like the Unix yes command without the loop.
func (*RootCmd) yes() *serpent.Command {
	return &serpent.Command{
		Use:   "yes",
		Short: "Print yes",
		Middleware: serpent.Chain(
			serpent.RequireNArgs(0),
		),
		Handler: func(inv *serpent.Invocation) error {
			_, err := fmt.Fprintln(inv.Stdout, "yes")
			return err
		},
	}
}
