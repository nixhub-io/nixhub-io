package css

import (
	"strings"
	"testing"
)

func TestBuildCSS(t *testing.T) {
	var rules = Rules{{
		Selector: "body > main#id",
		Properties: [][2]string{
			{"display", "flex"},
			{"content", Escape("joe")},
		},
	}, {
		Selector: "h1",
		Properties: [][2]string{
			{"margin", "0"},
		},
	}}

	var results = []string{
		"body > main#id {display:flex;content:\"joe\";}",
		"h1 {margin:0;}",
		"", // trailing new line
	}

	for i, r := range strings.Split(rules.CSS(), "\n") {
		if r != results[i] {
			t.Fatalf("CSS at line %d failed %s:\n%s\n%s",
				i, "(top expected, bottom returned)",
				results[i], r)
		}
	}
}

func TestEscape(t *testing.T) {
	var test = `
		biggest think in the big wild west" SIKES '` + "\U0001F914"
	var results = `"\0a \09 \09 biggest think in the big wild west\22  SIKES \27 \01f914 "`

	if esc := string(Escape(test)); esc != results {
		t.Fatalf("Escape returns unexpected result %s:\n%s\n%s",
			"(top expected, bottom returned)", results, esc)
	}
}

func TestInternalEscape(t *testing.T) {
	var buffer = make([]rune, 0, 6)

	var emoji = '\U0001F914'

	if esc := string(escape(emoji, buffer)); esc != "01f914" {
		t.Fatal("Escape returns unexpected result:", esc)
	}
}
