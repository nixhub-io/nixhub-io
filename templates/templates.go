package templates

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/phogolabs/parcello"
	"github.com/pkg/errors"
	"gitlab.com/shihoya-inc/errchi"
)

//go:generate go run github.com/phogolabs/parcello/cmd/parcello -r -i *.go

var Frontpage = template.New("").Funcs(htmlFns)

func Initialize() {
	var (
		parts = []parcello.File{
			mustResolve("tmpl_homepage.html"),
			mustResolve("tmpl_messages.html"),
			mustResolve("tmpl_message.html"),
		}
	)

	if err := parseFiles(Frontpage, parts...); err != nil {
		log.Panicln(err)
	}
}

type Renderer interface {
	Render(w io.Writer) error
}

type RendererWithTimezone interface {
	RenderWithTimezone(w io.Writer, tz *time.Location) error
}

func QuickRender(content Renderer) errchi.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		return Render(w, content, nil)
	}
}

func Render(w io.Writer, content Renderer, tz *time.Location) (int, error) {
	var err error

	switch content := content.(type) {
	case RendererWithTimezone:
		if tz == nil {
			tz = time.Local
		}
		err = content.RenderWithTimezone(w, tz)
	case nil:
		err = Frontpage.Execute(w, nil)
	default:
		err = content.Render(w)
	}

	if err != nil {
		return 500, err
	}

	return 200, nil
}

func parseFiles(tmpl *template.Template, files ...parcello.File) error {
	for _, f := range files {
		s, err := f.Stat()
		if err != nil {
			return errors.Wrap(err, "Failed to stat")
		}

		b, err := ioutil.ReadAll(f)
		if err != nil {
			return errors.Wrap(err, "Failed to read "+s.Name())
		}

		if _, err := tmpl.Parse(string(b)); err != nil {
			return errors.Wrap(err, "Failed to parse "+s.Name())
		}
	}

	return nil
}

func mustResolve(file string) parcello.File {
	f, err := parcello.Open(file)
	if err != nil {
		log.Fatalln("Failed to resolve " + file)
	}

	return f
}
