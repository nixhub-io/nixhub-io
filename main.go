package main

import (
	"log"
	"net/http"
	"os"

	"git.sr.ht/~diamondburned/gocad"
	"github.com/bwmarrin/discordgo"
	"gitlab.com/nixhub/nixhub.io/templates"
	"gitlab.com/shihoya-inc/errchi"
)

func main() {
	// nixhubd
	d, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Println("AAAA:", err)
	}

	d.State.TrackChannels = true
	d.State.TrackEmojis = true
	d.State.TrackMembers = true
	d.State.TrackRoles = true
	d.State.TrackVoice = false

	templates.Session = d

	r := errchi.NewRouter()
	r.Get("/", func(w http.ResponseWriter, r *http.Request) (int, error) {
		if err := templates.RenderHomepage(w, ChatLog); err != nil {
			return 500, err
		}

		return 200, nil
	})

	if err := d.Open(); err != nil {
		log.Println("Failed to connect to Discord:", err)
	}

	defer d.Close()

	if err := gocad.Serve(":8080", r); err != nil {
		log.Fatalln("Failed to start gocad:", err)
	}
}
