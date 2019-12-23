package discord

import (
	"errors"

	"github.com/bwmarrin/discordgo"
)

func Member(s *discordgo.Session, guildID, memberID string) (*discordgo.Member, error) {
	m, err := s.State.Member(guildID, memberID)
	if err != nil {
		m, err = s.GuildMember(guildID, memberID)
		if err != nil {
			return nil, err
		}

		s.State.MemberAdd(m)
	}

	return m, nil
}

func Role(s *discordgo.Session, guildID, roleID string) (*discordgo.Role, error) {
	r, err := s.State.Role(guildID, roleID)
	if err != nil {
		roles, err := s.GuildRoles(guildID)
		if err != nil {
			return nil, err
		}

		for _, role := range roles {
			s.State.RoleAdd(guildID, role)

			if role.ID == roleID {
				r = role
			}
		}
	}

	if r == nil {
		return nil, errors.New("role not found")
	}

	return r, nil
}

func Channel(s *discordgo.Session, channelID string) (*discordgo.Channel, error) {
	c, err := s.State.Channel(channelID)
	if err != nil {
		c, err = s.Channel(channelID)
		if err != nil {
			return nil, err
		}

		s.State.ChannelAdd(c)
	}

	return c, nil
}
