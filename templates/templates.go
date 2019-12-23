package templates

import (
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/bwmarrin/discordgo"
	"github.com/markbates/pkger"
	"github.com/markbates/pkger/pkging"
	"github.com/pkg/errors"
	"gitlab.com/shihoya-inc/errchi"
)

var Frontpage = template.New("").Funcs(htmlFns)
var Session *discordgo.Session

func init() {
	pkger.Include("/templates/")
	pkger.Include("/static/")

	var (
		parts = []pkging.File{
			mustResolve("/templates/tmpl_homepage.html"),
			mustResolve("/templates/tmpl_messages.html"),
			mustResolve("/templates/tmpl_message.html"),
		}

		// Check must exist
		_ = mustResolve("/static/style.css")
	)

	if err := parseFiles(Frontpage, parts...); err != nil {
		log.Panicln(err)
	}
}

func MountDir(path string) (string, http.Handler) {
	return path + "/",
		http.StripPrefix(path+"/", http.FileServer(pkger.Dir(path)))
}

type Renderer interface {
	Render(w io.Writer) error
}

func QuickRender(content Renderer) errchi.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) (int, error) {
		return RenderHomepage(w, content)
	}
}

func RenderHomepage(w io.Writer, content Renderer) (int, error) {
	var err error

	if content == nil {
		err = Frontpage.Execute(w, nil)
	} else {
		err = content.Render(w)
	}

	if err != nil {
		return 500, err
	}

	return 200, nil
}

func parseFiles(tmpl *template.Template, files ...pkging.File) error {
	for _, f := range files {
		b, err := ioutil.ReadAll(f)
		if err != nil {
			return errors.Wrap(err, "Failed to read "+f.Name())
		}

		if _, err := tmpl.Parse(string(b)); err != nil {
			return errors.Wrap(err, "Failed to parse "+f.Name())
		}
	}

	return nil
}

func mustResolve(file string) pkging.File {
	f, err := pkger.Open(file)
	if err != nil {
		log.Fatalln("Failed to resolve " + file)
	}

	return f
}
