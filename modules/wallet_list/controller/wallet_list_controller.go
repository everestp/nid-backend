package controller

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"

	"nid-backend/modules/wallet_list/dto"
	"nid-backend/modules/wallet_list/service"
	"nid-backend/pkg/middleware"
)

type WalletController struct {
	service *service.WalletService
}

func NewwalletListtController(
	service *service.WalletService,
) *WalletController {

	return &WalletController{
		service: service,
	}
}

// ============================================================
// USER ID FROM TOKEN
// ============================================================

func getUserID(
	r *http.Request,
) (string, error) {

	userID, ok := r.Context().
		Value("user_id").
		(string)

	if !ok || userID == "" {
		return "", errors.New(
			"unauthorized",
		)
	}

	return userID, nil
}

// ============================================================
// CREATE
// POST /wallets
// ============================================================

func (c *WalletController) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
// --------------------------------------------------------
	// Get user ID from authenticated session
	// --------------------------------------------------------

	userID, ok := r.Context().Value(
		middleware.UserIDKey,
	).(string)

	if !ok || userID == "" {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	var req dto.CreateWalletRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {

		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]interface{}{
				"success": false,
				"message": "invalid request body",
			},
		)

		return
	}

	wallet, err := c.service.Create(
		userID,
		req,
	)

	if err != nil {

		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]interface{}{
				"success": false,
				"message": err.Error(),
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		map[string]interface{}{
			"success": true,
			"message": "wallet added successfully",
			"wallet":  wallet,
		},
	)
}

// ============================================================
// GET ALL
// GET /wallets
// ============================================================

func (c *WalletController) GetAll(
	w http.ResponseWriter,
	r *http.Request,
) {

		// --------------------------------------------------------
	// Get user ID from authenticated session
	// --------------------------------------------------------

	userID, ok := r.Context().Value(
		middleware.UserIDKey,
	).(string)

	if !ok || userID == "" {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	wallets, err := c.service.GetAll(
		userID,
	)

	if err != nil {

		writeJSON(
			w,
			http.StatusInternalServerError,
			map[string]interface{}{
				"success": false,
				"message": err.Error(),
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"wallets": wallets,
		},
	)
}

// ============================================================
// GET BY ID
// GET /wallets/{id}
// ============================================================

func (c *WalletController) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {

	// --------------------------------------------------------
	// Get user ID from authenticated session
	// --------------------------------------------------------

	userID, ok := r.Context().Value(
		middleware.UserIDKey,
	).(string)

	if !ok || userID == "" {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}


	id := chi.URLParam(
		r,
		"id",
	)

	wallet, err := c.service.GetByID(
		userID,
		id,
	)

	if err != nil {

		writeJSON(
			w,
			http.StatusNotFound,
			map[string]interface{}{
				"success": false,
				"message": err.Error(),
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"wallet":  wallet,
		},
	)
}

// ============================================================
// UPDATE
// PUT /wallets/{id}
// ============================================================

func (c *WalletController) Update(
	w http.ResponseWriter,
	r *http.Request,
) {

		// --------------------------------------------------------
	// Get user ID from authenticated session
	// --------------------------------------------------------

	userID, ok := r.Context().Value(
		middleware.UserIDKey,
	).(string)

	if !ok || userID == "" {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}


	id := chi.URLParam(
		r,
		"id",
	)

	var req dto.UpdateWalletRequest

	if err := json.NewDecoder(
		r.Body,
	).Decode(&req); err != nil {

		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]interface{}{
				"success": false,
				"message": "invalid request body",
			},
		)

		return
	}

	wallet, err := c.service.Update(
		userID,
		id,
		req,
	)

	if err != nil {

		writeJSON(
			w,
			http.StatusBadRequest,
			map[string]interface{}{
				"success": false,
				"message": err.Error(),
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"message": "wallet updated successfully",
			"wallet":  wallet,
		},
	)
}

// ============================================================
// DELETE
// DELETE /wallets/{id}
// ============================================================

func (c *WalletController) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	// --------------------------------------------------------
	// Get user ID from authenticated session
	// --------------------------------------------------------

	userID, ok := r.Context().Value(
		middleware.UserIDKey,
	).(string)

	if !ok || userID == "" {
		http.Error(
			w,
			"Unauthorized",
			http.StatusUnauthorized,
		)
		return
	}

	id := chi.URLParam(
		r,
		"id",
	)

	err := c.service.Delete(
		userID,
		id,
	)

	if err != nil {

		writeJSON(
			w,
			http.StatusNotFound,
			map[string]interface{}{
				"success": false,
				"message": err.Error(),
			},
		)

		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"message": "wallet deleted successfully",
		},
	)
}

// ============================================================
// JSON
// ============================================================

func writeJSON(
	w http.ResponseWriter,
	status int,
	data interface{},
) {

	w.Header().Set(
		"Content-Type",
		"application/json",
	)

	w.WriteHeader(status)

	_ = json.NewEncoder(
		w,
	).Encode(data)
}
