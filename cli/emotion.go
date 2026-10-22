package cli

import (
	"fmt"
	"strings"

	"github.com/coder/serpent"
)

const (
	// envEmotion is the environment variable holding the emotion a
	// session was started with.
	envEmotion = "CODER_EMOTION"
	// defaultEmotion is printed when no emotion is set.
	defaultEmotion = "ambivalent"
)

// emotion prints the emotion that started the session.
func (*RootCmd) emotion() *serpent.Command {
	return &serpent.Command{
		Use:   "emotion",
		Short: "Print the emotion that started the session",
		Long:  "Prints the value of " + envEmotion + ", or \"" + defaultEmotion + "\" if it is unset or empty.",
		Handler: func(inv *serpent.Invocation) error {
			emotion := strings.TrimSpace(inv.Environ.Get(envEmotion))
			if emotion == "" {
				emotion = defaultEmotion
			}
			_, err := fmt.Fprintln(inv.Stdout, emotion)
			return err
		},
	}
}
