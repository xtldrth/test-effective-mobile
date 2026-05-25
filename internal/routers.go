package internal

import "net/http"

func RegisterSubscriptionsRoutes(mux *http.ServeMux, handler SubscriptionsHandler) {
	mux.HandleFunc("GET /api/subscriptions", handler.Get)
	mux.HandleFunc("GET /api/subscriptions/sum/", handler.GetPricesSum)
	mux.HandleFunc("GET /api/subscriptions/{id}/", handler.GetByID)
	mux.HandleFunc("POST /api/subscriptions", handler.Create)
	mux.HandleFunc("DELETE /api/subscriptions/{id}", handler.Delete)
	mux.HandleFunc("PATCH /api/subscriptions/{id}", handler.Update)
}
