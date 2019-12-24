package templates

import (
	"fmt"
	"io"
)

type MessageDelete struct {
	ID string
}

func (d *MessageDelete) Render(w io.Writer) error {
	_, err := fmt.Fprintf(w, `
	<style type="text/css">
		div.message[id="%s"] { display: none; }
	</style>`, d.ID)

	return err
}
