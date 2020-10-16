package templates

import (
	"io"

	"github.com/diamondburned/arikawa/discord"
	"github.com/nixhub-io/nixhub-io/css"
)

type MessageDelete struct {
	ID discord.MessageID
}

func (d MessageDelete) Render(w io.Writer) error {
	return css.WrapHTMLTo(w, css.Single(
		"div.message[id='"+d.ID.String()+"']", "display", "none",
	).CSS())
}
