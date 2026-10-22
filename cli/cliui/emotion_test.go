package cliui_test

import (
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/cli/cliui"
)

func TestApplyEmotion(t *testing.T) {
	// Mutates the package-level DefaultStyles, so not parallel.
	orig := cliui.DefaultStyles
	t.Cleanup(func() { cliui.DefaultStyles = orig })

	for _, emotion := range []string{"", cliui.EmotionHappy, cliui.EmotionSad, "Happy"} {
		require.NoError(t, cliui.ApplyEmotion(emotion), emotion)
	}
	require.Error(t, cliui.ApplyEmotion("angry"))
}
