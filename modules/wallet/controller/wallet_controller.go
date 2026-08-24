// modules/wallet/controller/wallet_controller.go
package controller

import (
	"encoding/json"
	"net/http"
	"nid-backend/modules/wallet/dto"
	"nid-backend/modules/wallet/service"
	"nid-backend/pkg/middleware"
)

type WalletController struct {
	service *service.WalletService
}

func NewWalletController(service *service.WalletService) *WalletController {
	return &WalletController{service: service}
}

func (c *WalletController) LinkWalletHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userID, ok := r.Context().Value(middleware.UserIDKey).(string)
	if !ok || userID == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var req dto.LinkWalletRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	wallet, err := c.service.LinkWallet(userID, req.Chain, req.Network, req.Address)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(dto.WalletResponse{
		ID:       wallet.ID,
		UserID:   wallet.UserID,
		Chain:    wallet.Chain,
		Network:  wallet.Network,
		Address:  wallet.Address,
		Status:   wallet.Status,
		LinkedAt: wallet.LinkedAt.Format("2006-01-02T15:04:05Z07:00"),
	})
}
