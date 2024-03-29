package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"

	_ "github.com/joho/godotenv/autoload"
)

func main() {

	c := make(chan os.Signal, 1)
	signal.Notify(c, syscall.SIGINT, syscall.SIGTERM)

	go bot()

	err := mqttInit()
	if err != nil {
		log.Fatal(err)
	}

	<-c // wait until a signal is received

	client.Disconnect(200)
	log.Println("end")
}
