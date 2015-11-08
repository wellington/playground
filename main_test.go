package main

import (
	"encoding/json"
	"strings"
	"testing"
)

type Response struct {
	Contents string
}

func TestCompile(t *testing.T) {
	in := strings.NewReader(`div { p { color: red; } }`)

	buf := mustCompile(doCompile(in))

	var resp Response
	err := json.Unmarshal(buf.Bytes(), &resp)
	if err != nil {
		t.Fatal(err)
	}

	if len(resp.Contents) == 0 {
		t.Fatal("empty response from server:\n", buf.String())
	}

	e := `/* line 1, stdin */
div p {
  color: red; }
`

	if e != resp.Contents {
		t.Errorf("received invalid contents\n% #v\n", resp.Contents)
	}
}
