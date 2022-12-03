package testutils

import (
	"bytes"
	"testing"

	"gotest.tools/v3/assert"
)

func TestFilter(t *testing.T) {
	cases := []string{
		"identitytoken",
		"identityToken",
		"IdentityToken",
		"password",
		"Password",
		"auth",
		"Auth",
	}

	for _, c := range cases {
		buf := bytes.NewBuffer([]byte(`{"` + c + `": "foo"}`))
		assert.Check(t, jsonIdentityTokenRegex.Match(buf.Bytes()))

		filtered := filterBuf(buf)
		assert.Check(t, bytes.Contains(filtered.Bytes(), []byte(`"<REDACTED>"`)))
		assert.Check(t, !bytes.Contains(filtered.Bytes(), []byte(`"foo"`)))
	}
}
