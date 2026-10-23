package cli_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/cli/clitest"
)

func TestEmotion(t *testing.T) {
	t.Parallel()

	for _, tt := range []struct {
		name     string
		args     []string
		expected string
		err      string
	}{
		{name: "Happy", args: []string{"emotion", "happy"}, expected: "Coder is happy! :D\n"},
		{name: "Sad", args: []string{"emotion", "sad"}, expected: "Coder is sad. :(\n"},
		{name: "Perturbed", args: []string{"emotion", "perturbed"}, expected: "Coder is perturbed. :/\n"},
		{name: "Delighted", args: []string{"emotion", "delighted"}, expected: "Coder is delighted! :D\n"},
		{name: "CaseInsensitive", args: []string{"emotion", "HAPPY"}, expected: "Coder is happy! :D\n"},
		{name: "Unknown", args: []string{"emotion", "angry"}, err: "unknown emotion"},
		{name: "MissingArg", args: []string{"emotion"}, err: "wanted 1 args"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inv, _ := clitest.New(t, tt.args...)
			buf := new(bytes.Buffer)
			inv.Stdout = buf
			inv.Stderr = buf
			err := inv.Run()
			if tt.err != "" {
				require.ErrorContains(t, err, tt.err)
				return
			}
			require.NoError(t, err)
			require.Equal(t, tt.expected, buf.String())
		})
	}
}
