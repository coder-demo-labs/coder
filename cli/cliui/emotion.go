package cliui

import (
	"strings"

	"golang.org/x/xerrors"

	"github.com/coder/pretty"
)

// Supported emotions for ApplyEmotion.
const (
	EmotionHappy = "happy"
	EmotionSad   = "sad"
)

// ApplyEmotion changes the theme of DefaultStyles to match the emotion. An
// empty emotion leaves the default theme untouched.
func ApplyEmotion(emotion string) error {
	switch strings.ToLower(strings.TrimSpace(emotion)) {
	case "":
		return nil
	case EmotionHappy:
		DefaultStyles.Keyword = pretty.Style{pretty.FgColor(Color("11"))}
		DefaultStyles.Placeholder = pretty.Style{pretty.FgColor(brightMagenta)}
		DefaultStyles.Fuchsia = pretty.Style{pretty.FgColor(brightMagenta)}
		DefaultStyles.DateTimeStamp = pretty.Style{pretty.FgColor(green)}
		DefaultStyles.FocusedPrompt = pretty.Style{
			pretty.FgColor(white),
			pretty.Wrap("> ", ""),
			pretty.FgColor(Color("11")),
		}
	case EmotionSad:
		DefaultStyles.Keyword = pretty.Style{pretty.FgColor(Color("4"))}
		DefaultStyles.Placeholder = pretty.Style{pretty.FgColor(Color("8"))}
		DefaultStyles.Fuchsia = pretty.Style{pretty.FgColor(Color("8"))}
		DefaultStyles.DateTimeStamp = pretty.Style{pretty.FgColor(Color("4"))}
		DefaultStyles.FocusedPrompt = pretty.Style{
			pretty.FgColor(white),
			pretty.Wrap("> ", ""),
			pretty.FgColor(Color("4")),
		}
	default:
		return xerrors.Errorf("unknown emotion %q: must be one of %q or %q", emotion, EmotionHappy, EmotionSad)
	}
	return nil
}
