package web_test

import (
	"io/fs"
	"testing"

	"github.com/OutOfStack/game-library-auth/internal/web"
	"github.com/stretchr/testify/require"
)

func TestTemplates(t *testing.T) {
	names, err := fs.Glob(web.Templates(), "*.html")
	require.NoError(t, err)
	require.ElementsMatch(t, []string{"unsubscribe.html", "unsubscribe_already.html", "unsubscribe_success.html"}, names)
}
