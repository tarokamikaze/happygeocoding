package main

import (
	"google.golang.org/appengine"
	"net/http"
	"github.com/tarokamikaze/happygeocoding/logic"
)

func main() {
	http.HandleFunc("/", logic.GeoHandler)
	appengine.Main()
}
