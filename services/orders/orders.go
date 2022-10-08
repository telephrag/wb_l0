package orders

import (
	"l0/cache"

	"github.com/jackc/pgx/v4/pgxpool"
)

type OrdersService struct {
	dbConn *pgxpool.Conn
	cache  *cache.OrdersCache
}
