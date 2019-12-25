package dispenser

import (
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/pkg/errors"
	"gitlab.com/nixhub/nixhub.io/discord"
	"gitlab.com/nixhub/nixhub.io/templates"
)

const TypingTimeout = 10 * time.Second

func (s *State) subscribeTyping() error {
	subEv := map[string]interface{}{
		"guild_id":   s.GuildID,
		"typing":     true,
		"activities": true,
	}

	if err := s.Session.SendWSEvent(14, subEv); err != nil {
		return errors.Wrap(err, "WS failed")
	}

	go s.typingLoop()

	s.addHandler(func(_ *discordgo.Session, t *discordgo.TypingStart) {
		if t.ChannelID != s.ChannelID {
			return
		}

		s.Typing <- t
	})

	return nil
}

func (s *State) StopTyping(user *discordgo.User) {
	s.StopType <- user
}

func (s *State) typingLoop() {
	s.Typing = make(chan *discordgo.TypingStart)
	s.Typers = &templates.Typing{
		Typers: make([]templates.Typer, 0, 5),
	}

	var ticker = time.NewTicker(time.Second)
	defer ticker.Stop()

	var stopper = make(chan *discordgo.User)
	s.StopType = stopper
	defer close(stopper)

	for {
		select {
		case t, ok := <-s.Typing:
			if !ok {
				return
			}

			m, err := discord.Member(s.Session, s.GuildID, t.UserID)
			if err != nil {
				// Shouldn't happen, we just ignore it
				continue
			}

			var typer = templates.Typer{
				UserID: m.User.ID,
				Name:   m.User.Username,
				When:   time.Unix(int64(t.Timestamp), 0),
			}

			if m.Nick != "" {
				typer.Name = m.Nick
			}

			s.Typers.AddTyper(typer)

		case user := <-stopper:
			if !s.Typers.Filter(func(t templates.Typer) bool {
				return user.ID == t.UserID
			}) {
				// Not changed, continue
				continue
			}

		case now := <-ticker.C:
			if !s.Typers.Filter(func(t templates.Typer) bool {
				// After means it's not timed out yet, so we keep it (true)
				return t.When.Add(TypingTimeout).After(now)
			}) {
				// Not changed, continue
				continue
			}
		}

		// All of the above mutate something, so we broadcast it. They don't if
		// they continue, though.
		go s.Broadcast(s.Typers)
	}
}
