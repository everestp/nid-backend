// modules/resolution/controller/resolution_controller.go
package controller

import (
	"encoding/json"
	"net/http"
	"nid-backend/modules/resolution/service"
)

type ResolutionController struct {
	service *service.ResolutionService
}

func NewResolutionController(service *service.ResolutionService) *ResolutionController {
	return &ResolutionController{service: service}
}

func (c *ResolutionController) ResolveHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	handle := r.URL.Query().Get("handle")
	chain := r.URL.Query().Get("chain")

	if handle == "" || chain == "" {
		http.Error(w, "Missing handle or chain query parameter", http.StatusBadRequest)
		return
	}

	address, err := c.service.Resolve(handle, chain)
	if err != nil {
		http.Error(w, err.Error(), http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"handle":  handle,
		"chain":   chain,
		"address": address,
	})
}
