package discord

import(
	"github.com/bwmarrin/discordgo"
	"fmt"
)

func Server(token string) (*discordgo.Session, error){
	dg, err := discordgo.New("Bot " + token)
	dg.AddHandler(messageListener)
    if err != nil {
        fmt.Println("error creating Discord session,", err)
        return nil, err
    }
	return dg, err

}

func messageListener(s *discordgo.Session, m *discordgo.MessageCreate){
	if m.Author.ID == s.State.User.ID {
        return
    }
	fmt.Println(m.Content)
}