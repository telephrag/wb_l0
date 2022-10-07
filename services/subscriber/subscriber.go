package subscriber

import (
	"context"
	"encoding/json"
	"fmt"
	"l0/services/model"
	"log"

	"github.com/go-playground/validator/v10"
	"github.com/jackc/pgtype"
	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/nats-io/stan.go"
)

type SubscriberService struct {
	StanConn    stan.Conn
	StanSubject string
}

func New(sc stan.Conn, stanSubject string) *SubscriberService {
	return &SubscriberService{
		StanConn:    sc,
		StanSubject: stanSubject,
	}
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

func insertIntoDB(
	ctx context.Context,
	conn *pgxpool.Conn,
	payload []byte,
	trackNumber, uid string,
) error {
	plJsonb := &pgtype.JSONB{Bytes: payload, Status: pgtype.Present}
	rows, err := conn.Query(
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

func (s *SubscriberService) Run(ctx context.Context, cancel context.CancelFunc, dbConn *pgxpool.Conn) (stan.Subscription, error) {

	sub, err := s.StanConn.Subscribe(s.StanSubject, func(msg *stan.Msg) {
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
		log.Printf("INFO: Order received: %s\n", o.OrderUID)

		if err := insertIntoDB(ctx, dbConn, msg.Data, o.OrderUID, o.TrackNumber); err != nil {
			log.Printf("%v\n", err)
			return
		}

	}, stan.MaxInflight(10), stan.AckWait(ACKWAIT_DURATION))
	if err != nil {
		return nil, err
	}

	return sub, nil
}
