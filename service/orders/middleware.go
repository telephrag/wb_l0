package orders

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/go-chi/chi"
)

func (s *OrdersService) GetOrder(next http.Handler) http.Handler {
	return http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		idParam := chi.URLParam(r, "id")

		o, ok := s.cache.Get(idParam)
		if !ok {
			log.Printf("INFO: Cache miss while retrieving order %s\n", idParam)

			row, err := s.dbConn.Query(
				r.Context(),
				fmt.Sprintf("select record from public.order where public.order.uid='%s'", idParam),
			)
			if err != nil {
				log.Print(err)
				// if pgErr, ok := err.(*pgconn.PgError); ok {
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(
					fmt.Sprintf("%d: Something went wrong querying for your order\n", http.StatusInternalServerError),
				))
				return
				// }
			}
			defer row.Close()

			if !row.Next() {
				log.Printf("ERROR: Attempt to retrieve a non-existant record %s\n", idParam)
				rw.WriteHeader(http.StatusBadRequest)
				rw.Write([]byte(
					fmt.Sprintf("%d: No record with provided id.", http.StatusBadRequest),
				))
				return
			}

			if err = row.Scan(&o); err != nil {
				log.Print(err)
				rw.WriteHeader(http.StatusInternalServerError)
				rw.Write([]byte(
					fmt.Sprintf("%d: %s\n", http.StatusInternalServerError, err.Error()),
				))
				return
			}
		}

		content, err := json.Marshal(o)
		if err != nil {
			log.Printf("ERROR: %v\n", err)
			rw.WriteHeader(http.StatusInternalServerError)
			rw.Write([]byte(
				fmt.Sprintf("%d: %s\n", http.StatusInternalServerError, err.Error()),
			))
		}

		rw.Header().Set("Content-type", "application/json")
		rw.Write(content)

		log.Printf("INFO: Successfully retrieved an order record %s\n", o.OrderUID)

		next.ServeHTTP(rw, r)
	})
}
