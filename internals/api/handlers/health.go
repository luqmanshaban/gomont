package handlers

import (
	"net/http"

	"github.com/luqmanshaban/gomont/internals/api/utils"
	"github.com/luqmanshaban/gomont/internals/store"
)

type HealthHandler struct {
	Store *store.HealthStore
}

func (h *HealthHandler) CheckDBStatus(w http.ResponseWriter, r *http.Request) {
	err := h.Store.CheckDBHealth()
	if err != nil {
		utils.WriteJson(w, http.StatusServiceUnavailable, map[string]string{"error": err.Error()})
	}
	utils.WriteJson(w, http.StatusOK, map[string]string{"message": "healthy"})
}
