// Package api wires the Huma HTTP API. Huma derives a JSON Schema from every
// registered operation's typed input and output, and validates every request
// against it by default, including rejecting unknown body fields
// (AllowAdditionalPropertiesByDefault is false), so nothing extra is needed
// per endpoint to get that validation.
package api

import (
	"net/http"

	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humago"
	"gorm.io/gorm"
)

type API struct {
	mux *http.ServeMux
	api huma.API
	db  *gorm.DB
}

func New(db *gorm.DB) *API {
	mux := http.NewServeMux()
	config := huma.DefaultConfig("Hikechievement API", "0.1.0")
	a := &API{mux: mux, api: humago.NewWithPrefix(mux, "/api", config), db: db}
	a.registerHealth()
	return a
}

func (a *API) Handler() http.Handler { return a.mux }
