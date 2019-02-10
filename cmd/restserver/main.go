package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/herdius/herdius-core/rpc/http/rest"
	"github.com/herdius/herdius-core/supervisor/transaction"
)

func main() {
	var txService transaction.Service

	txService = transaction.TxService()

	router := rest.Handler(txService)

	fmt.Println("Server is on tap now: http://localhost:8080/")
	log.Fatal(http.ListenAndServe(":8080", router))

}
