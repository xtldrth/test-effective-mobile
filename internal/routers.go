package internal

import "net/http"

func RegisterSubscriptionsRoutes(mux *http.ServeMux, handler SubscriptionsHandler) {
	mux.HandleFunc("GET /api/subscriptions", handler.Get)
	mux.HandleFunc("POST /api/subscriptions", handler.Create)
	mux.HandleFunc("PATCH /api/subscriptions", handler.Update)
	mux.HandleFunc("DELETE /api/subscriptions", handler.Delete)
	mux.HandleFunc("GET /api/subscriptions/sum/", handler.GetPricesSum)
	mux.HandleFunc("GET /api/subscriptions/{id}/", handler.GetByID)
}
