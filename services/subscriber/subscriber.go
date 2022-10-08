package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"l0/cache"
	"l0/services/model"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/stan.go"
)

type SubscriberService struct {
	stanConn    stan.Conn
	stanSubject string
	dbConn      *pgxpool.Conn
	cache       *cache.OrdersCache
}

func (s *SubscriberService) Init(
	stanConn stan.Conn,
	stanSubject string,
	dbConn *pgxpool.Conn,
	oc *cache.OrdersCache,
) (self *SubscriberService) {

	s.stanConn = stanConn
	s.stanSubject = stanSubject
	s.dbConn = dbConn
	s.cache = oc
	return s
}

func validatePayload(payload *model.Order) error {
	if err := validator.New().Struct(payload); err != nil {
		if vErr, ok := err.(validator.ValidationErrors); !ok {
			return fmt.Errorf("failed to cast error to type: %w", err)
		} else {
			return vErr
		}
	}

	return nil
}

func (s *SubscriberService) insertIntoDB(
	ctx context.Context,
	payload []byte,
	uid, trackNumber string,
) error {
	plJsonb := &pgtype.JSONB{Bytes: payload, Status: pgtype.Present}
	rows, err := s.dbConn.Query(
		ctx,
		"insert into public.order(uid, track_number, record) values ($1, $2, $3)",
		uid,
		trackNumber,
		plJsonb,
	)
	if err != nil {
		return err
	}

	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	return nil
}

func (s *SubscriberService) Run(ctx context.Context, cancel context.CancelFunc) (stan.Subscription, error) {

	sub, err := s.stanConn.Subscribe(s.stanSubject, func(msg *stan.Msg) {
		select {
		case <-ctx.Done():
			return
		default:
		}

		o := &model.Order{}
		if err := json.Unmarshal(msg.Data, &o); err != nil {
			log.Print(err)
			return
		}

		if err := validatePayload(o); err != nil {
			log.Printf("ERROR: %v\n", err)
			return
		}

		if err := s.insertIntoDB(ctx, msg.Data, o.OrderUID, o.TrackNumber); err != nil {
			log.Printf("%v: %s\n", err, o.OrderUID)
			return
		}

		if err := s.cache.Set(o.OrderUID, o); err != nil {
			log.Printf("ERROR: %v: %s\n", err, o.OrderUID)
			return
		}
		log.Printf("INFO: Order received: %s %v\n", o.OrderUID, o.DateCreated)

		log.Printf("Cache content: %d\n", s.cache.Len())
		s.cache.PrintIDs()
	}, // args
		stan.MaxInflight(10),
		stan.AckWait(ACKWAIT_DURATION),
		stan.DurableName(s.stanSubject+"_durable"),
	)
	if err != nil {
		return nil, err
	}

	return sub, nil
}
