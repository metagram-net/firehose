package wellknown

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
)

func Register(r *mux.Router) {
	r.HandleFunc("/health-check", func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, "OK")
	})
}
