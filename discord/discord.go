package discord

import(
	"github.com/bwmarrin/discordgo"
	"fmt"
)

func server() (*discordgo.Session, error){
	dg, err := discordgo.New("Bot " + Token)
    if err != nil {
        fmt.Println("error creating Discord session,", err)
        return
    }

}

func messageListener(s *discordgo.Session, m *discordgo.MessageCreate){
	if m.Author.ID == s.State.User.ID {
        return
    }
	fmt.Println(m.Content)
}