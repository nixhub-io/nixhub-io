package templates

import (
	"io"

	"gitlab.com/nixhub/nixhub.io/css"
)

type MessageDelete struct {
	ID string
}

func (d *MessageDelete) Render(w io.Writer) error {
	return css.WrapHTMLTo(w, css.Single(
		"div.message[id='"+d.ID+"']", "display", "none",
	).CSS())
}
