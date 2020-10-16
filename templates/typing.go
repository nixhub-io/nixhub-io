package templates

import (
	"io"
	"sync"
	"time"

	"github.com/diamondburned/arikawa/discord"
	"github.com/nixhub-io/nixhub-io/css"
)

type Typing struct {
	Typers []Typer
	lock   sync.RWMutex
}

type Typer struct {
	UserID discord.UserID
	Name   string
	When   time.Time
}

func (ts *Typing) AddTyper(t Typer) {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	for _, typer := range ts.Typers {
		if typer.UserID == t.UserID {
			typer.When = t.When
			return
		}
	}

	ts.Typers = append(ts.Typers, t)
}

// Filter filters out typers. A true returned keeps the struct. The returned
// bool is true when the slice is changed.
func (ts *Typing) Filter(fn func(Typer) bool) bool {
	ts.lock.Lock()
	defer ts.lock.Unlock()

	oldLen := len(ts.Typers)
	filtered := (ts.Typers)[:0]

	for _, t := range ts.Typers {
		if fn(t) {
			filtered = append(filtered, t)
		}
	}

	ts.Typers = filtered
	return len(filtered) != oldLen
}

func (ts *Typing) String() string {
	ts.lock.RLock()
	defer ts.lock.RUnlock()

	switch len(ts.Typers) {
	case 0:
		return ""
	case 1:
		return ts.Typers[0].Name
	case 2:
		return ts.Typers[0].Name + " and " + ts.Typers[1].Name
	default:
		var s string
		for i := 0; i < len(ts.Typers)-1; i++ {
			s += ts.Typers[i].Name + ", "
		}

		return s + " and " + ts.Typers[len(ts.Typers)-1].Name
	}
}

func (ts *Typing) Render(w io.Writer) error {
	if len(ts.Typers) == 0 {
		return css.WrapHTMLTo(w, css.Single(
			".typing", "visibility", "hidden").CSS())
	}

	rules := css.Rules{}
	rules.SetProperty(".typing", "visibility", "visible")

	// Check plurality
	if len(ts.Typers) == 1 {
		rules.SetProperty(".typing > .people::after",
			"content", `"is"`)
	} else {
		rules.SetProperty(".typing > .people::after",
			"content", `"are"`)
	}

	// Check number of people typing
	if len(ts.Typers) > 3 {
		rules.SetProperty(".typing > .people::before",
			"content", `""`)
		rules.SetProperty(".typing > .people::after",
			"content", `"People are"`)
	} else {
		rules.SetRule(css.Content(".typing > .people::before",
			ts.String()))
	}

	return css.WrapHTMLTo(w, rules.CSS())
}
