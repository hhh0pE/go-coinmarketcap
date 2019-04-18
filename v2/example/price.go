package main

import (
	"fmt"
	"log"

	cmc "github.com/hhh0pE/go-coinmarketcap/v2"
)

func main() {
	// Get info about coin
	price, err := cmc.Price(&cmc.PriceOptions{
		Symbol:  "ETH",
		Convert: "EUR",
	})
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(price)
}
