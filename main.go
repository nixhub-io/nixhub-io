package main

import (
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"git.sr.ht/~diamondburned/gocad"
	"github.com/diamondburned/arikawa/v2/discord"
	"github.com/diamondburned/arikawa/v2/gateway"
	"github.com/diamondburned/arikawa/v2/state"
	"github.com/diamondburned/arikawa/v2/state/store"
	"github.com/diamondburned/arikawa/v2/state/store/defaultstore"
	"github.com/nixhub-io/nixhub-io/dispenser"
	"github.com/nixhub-io/nixhub-io/static"
	"github.com/nixhub-io/nixhub-io/templates"
	"gitlab.com/shihoya-inc/errchi"

	_ "github.com/nixhub-io/nixhub-io/static"
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

	noopCab := store.NoopCabinet
	noopCab.RoleStore = defaultstore.NewRole()
	noopCab.EmojiStore = defaultstore.NewEmoji()
	noopCab.MemberStore = defaultstore.NewMember()
	noopCab.MessageStore = defaultstore.NewMessage(35)

	start("Discord", func() {
		// nixhubd
		s, err = state.NewWithStore("Bot "+token, noopCab)
		if err != nil {
			log.Fatalln("AAAA:", err)
		}

		s.Gateway.AddIntents(gateway.IntentGuilds)
		s.Gateway.AddIntents(gateway.IntentGuildEmojis)
		s.Gateway.AddIntents(gateway.IntentGuildMembers)
		s.Gateway.AddIntents(gateway.IntentGuildMessages)
		s.Gateway.AddIntents(gateway.IntentGuildMessageTyping)

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
	r.Mount("/static", http.StripPrefix("/static", static.Handler))
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
