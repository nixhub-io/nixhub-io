package main

import (
	"fmt"
	"log"
	"os"

	"github.com/diamondburned/arikawa/discord"
	"github.com/diamondburned/arikawa/session"
	"github.com/k0kubun/pp"
)

const header = `
package main

import "github.com/diamondburned/arikawa/discord"

`

func main() {
	if len(os.Args) != 3 {
		log.Fatalln(os.Args[0], "<token>", "<channelID>")
	}

	d, err := session.New(os.Args[1])
	if err != nil {
		log.Fatalln("Failed to connect:", err)
	}

	c, err := discord.ParseSnowflake(os.Args[2])
	if err != nil {
		log.Fatalln("Failed to parse snowflake:", err)
	}

	msgs, err := d.Messages(discord.ChannelID(c), 10)
	if err != nil {
		log.Fatalln("Failed to get 10 messages:", err)
	}

	fmt.Println(header)

	pp.ColoringEnabled = false
	fmt.Println("var ChatLog =", pp.Sprint(msgs))
}
