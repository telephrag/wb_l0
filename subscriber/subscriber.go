package subscriber

import (
	"encoding/json"
	"l0/model"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/nats-io/stan.go"
)

func SubscribeToOrders() {
	sc, err := stan.Connect("test-cluster", "sub", stan.NatsURL("nats://localhost:8223"))
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	sc.Subscribe("orders", func(msg *stan.Msg) {
		o := &model.Order{}
		if err := json.Unmarshal(msg.Data, &o); err != nil {
			log.Println(err)
			return
		}

		if err = validator.New().Struct(o); err != nil {
			if vErr, ok := err.(validator.ValidationErrors); !ok {
				log.Printf("failed to cast error to validator's type: %v\n", err)
			} else {
				log.Println(vErr)
			}
			return
		}

		log.Printf("%+v\n", o)
	})
}
