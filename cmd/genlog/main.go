package main

import (
	"fmt"
	"log"
	"os"

	"github.com/bwmarrin/discordgo"
	"github.com/k0kubun/pp"
)

const header = `
package main

import "github.com/bwmarrin/discordgo"

`

func main() {
	if len(os.Args) != 3 {
		log.Fatalln(os.Args[0], "<token>", "<channelID>")
	}

	d, err := discordgo.New(os.Args[1])
	if err != nil {
		log.Fatalln("Failed to connect:", err)
	}

	msgs, err := d.ChannelMessages(os.Args[2], 10, "", "", "")
	if err != nil {
		log.Fatalln("Failed to get 10 messages:", err)
	}

	fmt.Println(header)

	pp.ColoringEnabled = false
	fmt.Println("var ChatLog =", pp.Sprint(msgs))
}
