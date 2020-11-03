package main

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/joho/godotenv"
	"github.com/zergu1ar/steam"
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	client, err := steam.NewClient(new(http.Client), "", &steam.Credentials{
		Username:       os.Getenv("username"),
		Password:       os.Getenv("password"),
		SharedSecret:   os.Getenv("sharedSecret"),
		IdentitySecret: os.Getenv("identitySecret"),
	})
	if err != nil {
		log.Fatal(err)
	}

	err = client.Login()
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("login success")

	key, err := client.GetWebAPIKey()
	if err != nil {
		log.Fatal(err)
	}
	log.Print("Web Api Key: ", key)

	confirmations, err := client.GetConfirmations()
	if err != nil {
		log.Fatal(err)
	}

	for i := range confirmations {
		c := confirmations[i]
		log.Printf("Confirmation ID: %d, Key: %d\n", c.ID, c.Key)
		log.Printf("-> Title %s\n", c.Title)
		log.Printf("-> Receiving %s\n", c.Receiving)
		log.Printf("-> Since %s\n", c.Since)
		log.Printf("-> OfferID %d\n", c.OfferID)

		err = client.AnswerConfirmation(c, steam.AnswerDeny)
		if err != nil {
			log.Fatal(err)
		}

		log.Printf("Declined %d\n", c.ID)
	}

	log.Println("Bye!")
}
