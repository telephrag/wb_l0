package main

import (
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"log"
	"sync/atomic"
	"time"

	stan "github.com/nats-io/stan.go"
)

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	sc, err := stan.Connect("test-cluster", "pub", stan.NatsURL("nats://localhost:8223"))
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	for i := 0; i < 3; i++ {

		time.Sleep(PUBLISH_FREQUENCY_MICROSECONDS)

		o := baseOrder
		b := make([]byte, 8)
		binary.LittleEndian.PutUint64(b, ordersGenerated)
		atomic.AddUint64(&ordersGenerated, 1)
		o.OrderUID = base64.StdEncoding.EncodeToString(b)
		o.Payment.Transaction = o.OrderUID
		o.TrackNumber = o.OrderUID + "_track"
		o.Items[0].TrackNumber = o.TrackNumber
		data, err := json.Marshal(baseOrder)
		if err != nil {
			log.Panic(err)
		}

		if _, err := sc.PublishAsync("orders", data, func(s string, err error) {
			if err == nil {
				log.Println(s, "no error")
			} else {
				log.Println(err)
			}
		}); err != nil {
			log.Print(err)
		}
	}

	if err := sc.Publish("orders", []byte("{lol}")); err != nil {
		log.Print(err)
	}
}
