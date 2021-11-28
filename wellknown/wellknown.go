package wellknown

import (
	"fmt"
	"net/http"
)

func HealthCheck(w http.ResponseWriter, _ *http.Request) {
	fmt.Fprint(w, "OK")
}
