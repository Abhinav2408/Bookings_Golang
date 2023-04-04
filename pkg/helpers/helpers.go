package helpers

import (
	"BookingProject/pkg/config"
	"fmt"
	"net/http"
	"runtime/debug"
)

var app *config.AppConfig

// sets appconfig for helpers
func NewHelpers(a *config.AppConfig) {
	app = a
}

//two types of errors, client induced and server induced

func ClientError(w http.ResponseWriter, status int) {
	app.InfoLog.Println("Client error with status of", status)
	http.Error(w, http.StatusText(status), status)
}

func ServerError(w http.ResponseWriter, err error) {
	trace := fmt.Sprintf("%s\n%s", err.Error(), debug.Stack())
	app.ErrorLog.Println(trace)
	http.Error(w, http.StatusText(http.StatusInternalServerError), http.StatusInternalServerError)

}

func IsAuthenticated(req *http.Request) bool {
	exists := app.Session.Exists(req.Context(), "user_id")

	return exists
}
