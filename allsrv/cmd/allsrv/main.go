package main

import (
	"log"
	"net/http"
	"os"

	"github.com/jsteenb2/mess/allsrv"
)

func main() {
	db := new(allsrv.InmemDB)

	var svr http.Handler
	switch os.Getenv("ALLSRV_SERVER") {
	case "v1":
		log.Println("starting v1 server")
		svr = allsrv.NewServer(db, allsrv.WithBasicAuth("admin", "pass"))
	case "v2":
		log.Println("starting v2 server")
		svr = allsrv.NewServerV2(db, allsrv.WithBasicAuthV2("admin", "pass"))
	default: // run both
		log.Println("starting combination v1/v2 server")
		mux := http.NewServeMux()
		allsrv.NewServer(db, allsrv.WithMux(mux), allsrv.WithBasicAuth("admin", "pass"))
		allsrv.NewServerV2(db, allsrv.WithMux(mux), allsrv.WithBasicAuthV2("admin", "pass"))
		svr = mux
	}

	port := ":8091"
	log.Println("listening at http://localhost" + port)
	if err := http.ListenAndServe(port, svr); err != nil && err != http.ErrServerClosed {
		log.Println(err.Error())
		os.Exit(1)
	}
}
