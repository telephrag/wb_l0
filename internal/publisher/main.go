package main

import (
	"encoding/json"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	stan "github.com/nats-io/stan.go"
)

func main() {
	sc, err := stan.Connect("test-cluster", "pub", stan.NatsURL("nats://localhost:8223"))
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	interupt := make(chan os.Signal, 1)
	signal.Notify(interupt, syscall.SIGTERM, syscall.SIGINT)

	for i := 0; i < 3; i++ {
		// for {
		select {
		case <-interupt:
			return
		default:
		}

		time.Sleep(PUBLISH_FREQUENCY_MICROSECONDS)

		data, err := json.Marshal(nextOrder())
		if err != nil {
			log.Panic(err)
		}

		sc.Publish("orders", data)
	}
}
