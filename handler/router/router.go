package router

import (
	"database/sql"
	"net/http"

	"github.com/TechBowl-japan/go-stations/handler"
	"github.com/TechBowl-japan/go-stations/service"
)

func NewRouter(todoDB *sql.DB) *http.ServeMux {
	// register routes
	mux := http.NewServeMux()
	svc := service.NewTODOService(todoDB)
	// * /todos エンドポイントを作成する
	mux.HandleFunc("/todos", handler.NewTODOHandler(svc).ServeHTTP)
	// * /healthz エンドポイントを作成する
	mux.HandleFunc("/healthz", handler.NewHealthzHandler().ServeHTTP)

	return mux
}
