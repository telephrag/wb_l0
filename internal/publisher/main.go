package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	stan "github.com/nats-io/stan.go"
)

var hash []byte

func init() {
	r := rand.New(rand.NewSource(time.Now().UnixMicro()))
	b := md5.Sum([]byte(fmt.Sprint(r.Int63())))
	hash = b[:]
}

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	sc, err := stan.Connect("test-cluster", "pub", stan.NatsURL("nats://localhost:8223"))
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	for i := 0; i < 3; i++ {

		time.Sleep(PUBLISH_FREQUENCY)

		o := baseOrder

		// creating new uid
		b := md5.Sum(hash)
		hash = b[:]
		o.OrderUID = base64.StdEncoding.EncodeToString(hash)
		o.Payment.Transaction = o.OrderUID
		o.TrackNumber = o.OrderUID + "_track"

		// timestamp cause cache uses it to determine which records to pull out
		o.DateCreated = time.Now()
		o.Items[0].TrackNumber = o.TrackNumber

		data, err := json.Marshal(baseOrder)
		if err != nil {
			log.Panic(err)
		}

		if _, err := sc.PublishAsync("orders", data, func(s string, err error) {
			if err == nil {
				log.Println("INFO: Published: ", o.OrderUID, o.DateCreated.UnixMicro())
			} else {
				log.Println(err)
			}
		}); err != nil {
			log.Println("ERROR: ", err)
		}
	}

	if err := sc.Publish("orders", []byte("{lol}")); err != nil {
		log.Println("ERROR: ", err)
	}
}
