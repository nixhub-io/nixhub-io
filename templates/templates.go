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
)

var Frontpage = template.New("").Funcs(htmlFns)
var Session *discordgo.Session

func init() {
	pkger.Include("/templates/")
	pkger.Include("/static/")

	var (
		parts = []pkging.File{
			mustResolve("/templates/tmpl_homepage.html"),
			mustResolve("/templates/tmpl_message.html"),
		}

		// Check must exist
		_ = mustResolve("/static/style.css")
		_ = mustResolve("/static/nfront.png")
	)

	if err := parseFiles(Frontpage, parts...); err != nil {
		log.Panicln(err)
	}
}

func MountDir(path string) (string, http.Handler) {
	return path + "/",
		http.StripPrefix(path+"/", http.FileServer(pkger.Dir(path)))
}

func RenderHomepage(w io.Writer, content interface{}) error {
	switch content := content.(type) {
	case []*discordgo.Message:
		return Frontpage.Execute(w, content)
	case *discordgo.Message:
		return Frontpage.ExecuteTemplate(w, "message", content)
	default:
		return errors.New("Unknown content type")
	}
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
