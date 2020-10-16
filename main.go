package main

import (
	"io/ioutil"
	"log"
	"os"
	"strings"
	"time"

	"git.sr.ht/~diamondburned/gocad"
	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/gateway"
	"github.com/diamondburned/arikawa/state"
	"github.com/nixhub-io/nixhub-io/dispenser"
	"github.com/nixhub-io/nixhub-io/store"
	"github.com/nixhub-io/nixhub-io/templates"
	"gitlab.com/shihoya-inc/errchi"
)

func main() {
	var token string
	var err error

	// Try and load a token from file
	if file := os.Getenv("TOKEN_FILE"); file != "" {
		f, err := ioutil.ReadFile(file)
		if err != nil {
			log.Fatalln("Failed to open", file+":", err)
		}

		token = strings.TrimSpace(string(f))
	} else {
		token = os.Getenv("BOT_TOKEN")
	}

	if token == "" {
		log.Fatalln("Token must not be empty!")
	}

	var s *state.State

	start("Discord", func() {
		// nixhubd
		s, err = state.NewWithStore("Bot "+token, store.New())
		if err != nil {
			log.Fatalln("AAAA:", err)
		}

		s.Gateway.AddIntent(gateway.IntentGuilds)
		s.Gateway.AddIntent(gateway.IntentGuildEmojis)
		s.Gateway.AddIntent(gateway.IntentGuildMembers)
		s.Gateway.AddIntent(gateway.IntentGuildMessages)
		s.Gateway.AddIntent(gateway.IntentGuildMessageTyping)

		if err := s.Open(); err != nil {
			log.Fatalln("Failed to connect to Discord:", err)
		}
	})

	defer s.Close()

	start("Templates", templates.Initialize)

	var d *dispenser.State

	start("Dispenser", func() {
		channelID, err := discord.ParseSnowflake(os.Getenv("CHANNEL_ID"))
		if err != nil {
			log.Fatalln("Failed to parse $CHANNEL_ID:", err)
		}

		d, err = dispenser.Initialize(s, discord.ChannelID(channelID))
		if err != nil {
			log.Fatalln("Failed to initialize dispenser:", err)
		}
	})

	r := errchi.NewRouter()
	r.Mount(templates.MountDir("/static"))
	r.Get("/feed", d.Handler)
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
