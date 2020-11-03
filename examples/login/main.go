package main

import (
	"fmt"
	"github.com/joho/godotenv"
	"github.com/zergu1ar/steam"
	"log"
	"net/http"
	"os"
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
}
