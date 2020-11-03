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

	client, err := steam.NewClient(new(http.Client), "", steam.LanguageRus, &steam.Credentials{
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

	sid := client.GetSteamId()
	apps, err := client.GetInventoryAppStats(sid)
	if err != nil {
		log.Fatal(err)
	}

	for _, v := range apps {
		log.Printf("-- AppID total asset count: %d\n", v.AssetCount)
		for _, context := range v.Contexts {
			log.Printf("-- Items on %d %d (count %d)\n", v.AppID, context.ID, context.AssetCount)
			inven, err := client.GetInventory(sid, v.AppID, context.ID, true)
			if err != nil {
				log.Fatal(err)
			}

			for _, item := range inven {
				log.Printf("Item: %s = %d\n", item.Desc.MarketHashName, item.AssetID)
			}
		}
	}

	log.Println("Bye!")
}
