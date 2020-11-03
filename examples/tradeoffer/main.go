package main

import (
	"fmt"
	"github.com/zergu1ar/steam"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/joho/godotenv"
)

func processOffer(c *steam.Client, offer *steam.TradeOffer) {
	var sid steam.SteamID
	sid.ParseDefaults(offer.Partner)

	log.Printf("Offer id: %d, Receipt ID: %d, State: %d", offer.ID, offer.ReceiptID, offer.State)
	log.Printf("Offer partner SteamID 64: %d", uint64(sid))
	if offer.State == steam.TradeStateAccepted {
		items, err := c.GetTradeReceivedItems(offer.ReceiptID)
		if err != nil {
			log.Printf("error getting items: %v", err)
		} else {
			for _, item := range items {
				log.Printf("Item: %#v", item)
			}
		}
	}
	if offer.State == steam.TradeStateActive && !offer.IsOurOffer {
		items, err := c.GetTradeReceivedItems(offer.ReceiptID)
		if err != nil {
			log.Printf("error getting items: %v", err)
			return
		}
		err = offer.Accept(c)
		if err != nil {
			log.Printf("error accept trade: %v", err)
			return
		}
		for _, item := range items {
			log.Printf("Item received: %#v", item)
		}
	}
}

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Fatal(err)
	}

	client, err := steam.NewClient(new(http.Client), "", "", &steam.Credentials{
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
	log.Print("Key: ", key)

	resp, err := client.GetTradeOffers(
		steam.TradeFilterSentOffers|steam.TradeFilterRecvOffers|steam.TradeFilterActiveOnly,
		time.Now(),
	)
	if err != nil {
		log.Fatal(err)
	}

	for _, offer := range resp.SentOffers {
		processOffer(client, offer)
	}
	for _, offer := range resp.ReceivedOffers {
		processOffer(client, offer)
	}

	log.Println("Bye!")
}
