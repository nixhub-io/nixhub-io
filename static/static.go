package static

import (
	"net/http"

	"github.com/phogolabs/parcello"
)

//go:generate go run github.com/phogolabs/parcello/cmd/parcello -r

var Handler = http.FileServer(parcello.Manager)
