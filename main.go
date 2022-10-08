package main

import (
	"context"
	"l0/cache"
	"l0/services/subscriber"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/stan.go"
)

func main() {
	log.SetFlags(log.Flags() &^ (log.Ldate | log.Ltime))

	sc, err := stan.Connect("test-cluster", "sub", stan.NatsURL("nats://localhost:8223"))
	if err != nil {
		log.Fatal(err)
	}
	defer sc.Close()

	pool, err := pgxpool.Connect(context.Background(), PSQL_DB_URL)
	if err != nil {
		log.Fatalf("Failed to connect to pgxPool: %v\n", err)
	}
	defer pool.Close()

	cache := (&cache.OrdersCache{}).Init(CACHE_CAPACITY)

	// Subscriber
	subCtx, subCancel := context.WithCancel(context.Background())

	subConn, err := pool.Acquire(subCtx)
	if err != nil {
		log.Fatalf("Failed to acquire conn from pool: %v\n", err)
	}
	defer subConn.Release()

	subscriber := (&subscriber.SubscriberService{}).Init(sc, "orders", subConn, cache)
	if sub, err := subscriber.Run(subCtx, subCancel); err != nil {
		// do I need to use durable subscription?
		// do I need to constantly check for connection with nats manually?
		// can safely log.Fatal() since we won't do anything else if subscription fails
		log.Fatal(err)
	} else {
		defer sub.Close()
	}
	// TODO: orders.Run()

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-interrupt

	subCancel()
}
