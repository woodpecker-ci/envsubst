package envsubst

import (
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/woodpecker-ci/envsubst/parse"
)

func TestParseError(t *testing.T) {
	tmpl, err := Parse("${unclosed")
	assert.ErrorIs(t, err, parse.ErrMissingClosingBrace)
	assert.Nil(t, tmpl)
}

func TestTemplateExecute(t *testing.T) {
	tmpl, err := Parse("${greeting:-hello} ${name}!")
	require.NoError(t, err)

	// a parsed template can be executed with different mappings
	got, err := tmpl.Execute(func(s string) string {
		return map[string]string{"name": "world"}[s]
	})
	require.NoError(t, err)
	assert.Equal(t, "hello world!", got)

	got, err = tmpl.Execute(func(s string) string {
		return map[string]string{"greeting": "hi", "name": "there"}[s]
	})
	require.NoError(t, err)
	assert.Equal(t, "hi there!", got)
}

func TestParseFile(t *testing.T) {
	dir := t.TempDir()

	t.Run("valid template", func(t *testing.T) {
		file := filepath.Join(dir, "valid.tmpl")
		require.NoError(t, os.WriteFile(file, []byte("image: ${IMAGE:-alpine}\n"), 0o600))

		tmpl, err := ParseFile(file)
		require.NoError(t, err)

		got, err := tmpl.Execute(func(string) string { return "" })
		require.NoError(t, err)
		assert.Equal(t, "image: alpine\n", got)
	})

	t.Run("missing file", func(t *testing.T) {
		tmpl, err := ParseFile(filepath.Join(dir, "missing.tmpl"))
		assert.ErrorIs(t, err, fs.ErrNotExist)
		assert.Nil(t, tmpl)
	})

	t.Run("invalid template", func(t *testing.T) {
		file := filepath.Join(dir, "invalid.tmpl")
		require.NoError(t, os.WriteFile(file, []byte("${unclosed"), 0o600))

		tmpl, err := ParseFile(file)
		assert.ErrorIs(t, err, parse.ErrMissingClosingBrace)
		assert.Nil(t, tmpl)
	})
}
