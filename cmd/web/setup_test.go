// this file runs before all our tests
package main

import (
	"net/http"
	"os"
	"testing"
)

func TestMain(m *testing.M) {

	os.Exit(m.Run())
}

type myHandler struct{}

// make same functions as the original http.handler has
func (mh *myHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {

}
