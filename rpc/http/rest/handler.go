package rest

import (
	"encoding/json"
	"net/http"

	"github.com/herdius/herdius-core/supervisor/transaction"
	"github.com/julienschmidt/httprouter"
)

// Handler ...
func Handler(t transaction.Service) http.Handler {
	router := httprouter.New()

	router.GET("/txlist", getTxList(t))
	router.POST("/tx", addTx(t))

	return router
}

// addTx returns a handler for POST /tx request
func addTx(t transaction.Service) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		decoder := json.NewDecoder(r.Body)
		var newTx transaction.Tx
		err := decoder.Decode(&newTx)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		t.AddTx(newTx)

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode("Raw transaction added")
	}
}

// GetTxList returns a handler for GET /txList requests
func getTxList(t transaction.Service) func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	return func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "application/json")
		txlist := t.GetTxList()
		json.NewEncoder(w).Encode(txlist)
	}
}
