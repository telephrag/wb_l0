package main

import (
	"log"

	stan "github.com/nats-io/stan.go"
)

func main() {
	sc, err := stan.Connect("test-cluster", "pub", stan.NatsURL("nats://localhost:8223"))
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	sc.Publish("hello", []byte("чырткем, мудила"))
}
