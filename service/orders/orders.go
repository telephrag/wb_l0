package orders

import (
	"context"
	"fmt"
	"l0/cache"
	"l0/service/model"
	"net/http"

	"github.com/go-chi/chi"
	"github.com/jackc/pgx/v4/pgxpool"
)

type OrdersService struct {
	dbConn   *pgxpool.Conn
	cache    *cache.OrdersCache
	endpoint string
}

func (s *OrdersService) restoreCache() error {
	ctx := context.Background()
	sql := fmt.Sprintf(
		"select * from public.order order by public.order.record->'date_created' desc limit %d",
		s.cache.Cap(),
	)
	rows, err := s.dbConn.Query(ctx, sql)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		o := model.Order{}
		if err := rows.Scan(nil, nil, &o); err != nil {
			return err
		}

		if rows.Err() != nil {
			return rows.Err()
		}

		s.cache.Set(o.OrderUID, &o)
	}
	rows.Close()
	if rows.Err() != nil {
		return rows.Err()
	}

	return nil
}

func (s *OrdersService) Init(
	dbConn *pgxpool.Conn,
	cache *cache.OrdersCache,
	endpoint string,
) (self *OrdersService, err error) {
	s.dbConn = dbConn
	s.cache = cache
	if err := s.restoreCache(); err != nil {
		return nil, err
	}
	s.endpoint = endpoint
	return s, nil
}

func (s *OrdersService) Run() http.Handler {
	s.cache.PrintIDs()
	r := chi.NewRouter()
	return r // TODO
}
