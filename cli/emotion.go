package cli

import (
	"fmt"
	"strings"

	"golang.org/x/xerrors"

	"github.com/coder/serpent"
)

// emotions maps each supported emotion to the greeting Coder starts with.
var emotions = map[string]string{
	"happy":     "Coder is happy! :D",
	"sad":       "Coder is sad. :(",
	"perturbed": "Coder is perturbed. :/",
	"delighted": "Coder is delighted! :D",
}

// emotion starts Coder as happy, sad, perturbed, or delighted.
func (*RootCmd) emotion() *serpent.Command {
	return &serpent.Command{
		Use:        "emotion <happy|sad|perturbed|delighted>",
		Short:      "Start Coder as happy, sad, perturbed, or delighted",
		Hidden:     true,
		Middleware: serpent.RequireNArgs(1),
		Handler: func(inv *serpent.Invocation) error {
			name := strings.ToLower(inv.Args[0])
			greeting, ok := emotions[name]
			if !ok {
				return xerrors.Errorf("unknown emotion %q: must be one of happy, sad, perturbed, delighted", inv.Args[0])
			}
			_, err := fmt.Fprintln(inv.Stdout, greeting)
			return err
		},
	}
}
