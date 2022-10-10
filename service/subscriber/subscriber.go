package subscriber

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"l0/cache"
	"l0/service/model"
	"log"
	"strings"
	"sync/atomic"
	"time"

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
	running     int32
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

var ErrInvalidOrderUID = errors.New("illegal symbols in order uid")

func validatePayload(payload *model.Order) error {
	if err := validator.New().Struct(payload); err != nil {
		if vErr, ok := err.(validator.ValidationErrors); !ok {
			return fmt.Errorf("failed to cast error to type: %w", err)
		} else {
			return vErr
		}
	}

	if strings.ContainsAny(payload.OrderUID, "{}/.,") {
		return ErrInvalidOrderUID
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
		atomic.AddInt32(&s.running, 1)
		defer atomic.AddInt32(&s.running, -1)

		o := &model.Order{}
		if err := json.Unmarshal(msg.Data, &o); err != nil {
			if errAck := msg.Ack(); errAck != nil {
				err = fmt.Errorf("%w: %s", err, errAck)
			}
			log.Printf("ERROR: %v\n", err)
			return
		}

		if err := validatePayload(o); err != nil {
			if errAck := msg.Ack(); errAck != nil {
				err = fmt.Errorf("%w: %s", err, errAck)
			}
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

		// log.Printf("Cache content: %d\n", s.cache.Len())
		// s.cache.PrintIDs()

		// No panic since subscription is durable
		// thus, we can try again after restart.
		if err := msg.Ack(); err != nil {
			log.Printf("ERROR: %s\n", err.Error())
		}
	}, // args
		stan.MaxInflight(10),
		stan.AckWait(ACKWAIT_DURATION),
		stan.DurableName(s.stanSubject+"_durable"),
		stan.SetManualAckMode(),
	)
	if err != nil {
		return nil, err
	}

	return sub, nil
}

func (s *SubscriberService) WaitUntilAllDone() {
	for atomic.LoadInt32(&s.running) != 0 {
		time.Sleep(time.Millisecond)
	}
}
