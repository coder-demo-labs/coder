package cli_test

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/cli/clitest"
)

func TestEmotion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		env      string
		expected string
	}{
		{name: "Unset", env: "", expected: "ambivalent\n"},
		{name: "Whitespace", env: "  ", expected: "ambivalent\n"},
		{name: "Set", env: "joyful", expected: "joyful\n"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			inv, _ := clitest.New(t, "emotion")
			if tt.env != "" {
				inv.Environ.Set("CODER_EMOTION", tt.env)
			}
			var out strings.Builder
			inv.Stdout = &out
			require.NoError(t, inv.Run())
			require.Equal(t, tt.expected, out.String())
		})
	}
}
