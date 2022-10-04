package main

import (
	"l0/subscriber"
	"log"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	go subscriber.SubscribeToOrders()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-interrupt
}
