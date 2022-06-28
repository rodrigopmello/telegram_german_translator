package main

import (
	"log"
	"net/http"
	"os"
	"telegram_german_translator/webhook"

	"github.com/joho/godotenv"
)

// The main funtion starts our server on a port specified in .env file
func main() {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println(".env file not found \n", err.Error())
		panic(err)
	}

	PORT := ":" + os.Getenv("PORT")
	log.Println("Starting web service")
	http.ListenAndServe(PORT, http.HandlerFunc(webhook.Handler))
}
