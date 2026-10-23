package cli_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/cli/clitest"
	"github.com/coder/coder/v2/testutil"
)

func TestYes(t *testing.T) {
	t.Parallel()

	inv, _ := clitest.New(t, "yes")
	buf := new(bytes.Buffer)
	inv.Stdout = buf
	require.NoError(t, inv.WithContext(testutil.Context(t, testutil.WaitShort)).Run())
	require.Equal(t, "yes\n", buf.String())
}
