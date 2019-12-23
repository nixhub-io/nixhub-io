package main

import (
	"log"
	"os"
	"time"

	"git.sr.ht/~diamondburned/gocad"
	"github.com/bwmarrin/discordgo"
	"gitlab.com/nixhub/nixhub.io/dispenser"
	"gitlab.com/nixhub/nixhub.io/templates"
	"gitlab.com/shihoya-inc/errchi"
)

func main() {
	// nixhubd
	d, err := discordgo.New("Bot " + os.Getenv("BOT_TOKEN"))
	if err != nil {
		log.Println("AAAA:", err)
	}

	d.StateEnabled = true
	d.State.TrackChannels = true
	d.State.TrackEmojis = true
	d.State.TrackMembers = true
	d.State.TrackRoles = true
	d.State.TrackVoice = false
	// dispenser keeps its own message pool
	d.State.MaxMessageCount = 0

	start("Discord", func() {
		if err := d.Open(); err != nil {
			log.Fatalln("Failed to connect to Discord:", err)
		}
	})

	defer d.Close()

	start("Dispenser", func() {
		if err := dispenser.Initialize(d, os.Getenv("CHANNEL_ID")); err != nil {
			log.Fatalln("Failed to initialize dispenser:", err)
		}
	})

	r := errchi.NewRouter()
	r.Mount(templates.MountDir("/static"))
	r.Get("/feed", dispenser.Handler)
	r.Get("/", templates.QuickRender(nil))

	log.Println("Serving at :8080")

	if err := gocad.Serve(":8080", r); err != nil {
		log.Fatalln("Failed to start gocad:", err)
	}
}

var startedWhen = map[string]time.Time{}

func start(thing string, fn func()) {
	log.Println("Starting", thing+"...")
	t := time.Now()

	fn()

	log.Println("Started", thing+",", "took", time.Now().Sub(t))
}
