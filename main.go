package main

import(
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"mc_bot/discord"
	"github.com/joho/godotenv"
)

func main(){
	err := godotenv.Load()
	token := os.Getenv("token")

	goSession, err := discord.Server(token)

	if err != nil{
		fmt.Println("Error has occured")
		return
	}

	err = goSession.Open()
    if err != nil {
        fmt.Println("error opening connection,", err)
        return
    }

    // Wait here until CTRL-C or other term signal is received.
    fmt.Println("Bot is now running. Press CTRL-C to exit.")

	defer goSession.Close()

	stop := make(chan os.Signal, 1)
signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
<-stop

}