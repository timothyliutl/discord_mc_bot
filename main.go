package main

import(
	"fmt"
	"discord"
)

func main(){

	goSession, err = discord.server()

	if err != nil{
		fmt.Println("Error has occured")
		return
	}

	defer goSession.Close()

}