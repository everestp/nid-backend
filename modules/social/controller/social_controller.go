package controller

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"

	"nid-backend/modules/social/dto"
	"nid-backend/modules/social/service"
	"nid-backend/pkg/middleware"
)

type SocialController struct {
	service *service.SocialService
}

func NewSocialController(
	service *service.SocialService,
) *SocialController {
	return &SocialController{
		service: service,
	}
}

// ============================================================
// GET /api/v1/social
// ============================================================
//
// Returns all social identities belonging to the authenticated
// user.
//
// ============================================================

func (c *SocialController) ListHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	identities, err := c.service.GetMySocials(userID)
	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"socials": identities,
			"count":   len(identities),
		},
	)
}

// ============================================================
// GET /api/v1/social/public
// ============================================================
//
// Returns only publicly visible social identities.
//
// ============================================================

func (c *SocialController) PublicListHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	identities, err := c.service.GetPublicSocials(userID)
	if err != nil {
		writeError(
			w,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"socials": identities,
			"count":   len(identities),
		},
	)
}

// ============================================================
// GET /api/v1/social/{id}
// ============================================================

func (c *SocialController) GetHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	id := strings.TrimSpace(
		r.PathValue("id"),
	)

	if id == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"social identity id is required",
		)
		return
	}

	identity, err := c.service.GetByID(
		userID,
		id,
	)

	if err != nil {
		if errors.Is(err, service.ErrSocialNotFound) {
			writeError(
				w,
				http.StatusNotFound,
				"social identity not found",
			)
			return
		}

		writeError(
			w,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"social":  identity,
		},
	)
}

// ============================================================
// POST /api/v1/social
// ============================================================

func (c *SocialController) CreateHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	var req dto.CreateSocialRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	req.Platform = strings.TrimSpace(
		req.Platform,
	)

	req.Handle = strings.TrimSpace(
		req.Handle,
	)

	if req.Platform == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"platform is required",
		)
		return
	}

	if req.Handle == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"handle is required",
		)
		return
	}

	identity, err := c.service.Create(
		userID,
		req,
	)

	if err != nil {
		if errors.Is(err, service.ErrInvalidPlatform) {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid social platform",
			)
			return
		}

		if errors.Is(err, service.ErrInvalidHandle) {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid social handle",
			)
			return
		}

		if errors.Is(err, service.ErrSocialExists) {
			writeError(
				w,
				http.StatusConflict,
				"social identity already exists",
			)
			return
		}

		writeError(
			w,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	writeJSON(
		w,
		http.StatusCreated,
		map[string]interface{}{
			"success": true,
			"social":  identity,
		},
	)
}

// ============================================================
// PUT /api/v1/social/{id}
// ============================================================

func (c *SocialController) UpdateHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	id := strings.TrimSpace(
		r.PathValue("id"),
	)

	if id == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"social identity id is required",
		)
		return
	}

	var req dto.UpdateSocialRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	if req.Handle != nil {
		value := strings.TrimSpace(*req.Handle)

		if value == "" {
			writeError(
				w,
				http.StatusBadRequest,
				"handle cannot be empty",
			)
			return
		}

		req.Handle = &value
	}

	identity, err := c.service.Update(
		userID,
		id,
		req,
	)

	if err != nil {
		if errors.Is(err, service.ErrSocialNotFound) {
			writeError(
				w,
				http.StatusNotFound,
				"social identity not found",
			)
			return
		}

		if errors.Is(err, service.ErrInvalidHandle) {
			writeError(
				w,
				http.StatusBadRequest,
				"invalid social handle",
			)
			return
		}

		if errors.Is(err, service.ErrSocialExists) {
			writeError(
				w,
				http.StatusConflict,
				"social identity already exists",
			)
			return
		}

		writeError(
			w,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"social":  identity,
		},
	)
}

// ============================================================
// PATCH /api/v1/social/{id}/visibility
// ============================================================
//
// Body:
//
// {
//     "publicly_visible": true
// }
//
// ============================================================

func (c *SocialController) ToggleVisibilityHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	id := strings.TrimSpace(
		r.PathValue("id"),
	)

	if id == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"social identity id is required",
		)
		return
	}

	var req dto.ToggleVisibilityRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(
			w,
			http.StatusBadRequest,
			"invalid request body",
		)
		return
	}

	identity, err := c.service.ToggleVisibility(
		userID,
		id,
		req.PubliclyVisible,
	)

	if err != nil {
		if errors.Is(err, service.ErrSocialNotFound) {
			writeError(
				w,
				http.StatusNotFound,
				"social identity not found",
			)
			return
		}

		writeError(
			w,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"social":  identity,
		},
	)
}

// ============================================================
// DELETE /api/v1/social/{id}
// ============================================================

func (c *SocialController) DeleteHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, err := getUserID(r)
	if err != nil {
		writeError(
			w,
			http.StatusUnauthorized,
			"unauthorized",
		)
		return
	}

	id := strings.TrimSpace(
		r.PathValue("id"),
	)

	if id == "" {
		writeError(
			w,
			http.StatusBadRequest,
			"social identity id is required",
		)
		return
	}

	err = c.service.Delete(
		userID,
		id,
	)

	if err != nil {
		if errors.Is(err, service.ErrSocialNotFound) {
			writeError(
				w,
				http.StatusNotFound,
				"social identity not found",
			)
			return
		}

		writeError(
			w,
			http.StatusInternalServerError,
			err.Error(),
		)
		return
	}

	writeJSON(
		w,
		http.StatusOK,
		map[string]interface{}{
			"success": true,
			"message": "social identity deleted successfully",
		},
	)
}

// ============================================================
// AUTHENTICATED USER ID
// ============================================================

func getUserID(r *http.Request) (string, error) {
	userID, ok := r.Context().Value(
		middleware.UserIDKey,
	).(string)

	if !ok || strings.TrimSpace(userID) == "" {
		return "", errors.New("user id not found")
	}

	return strings.TrimSpace(userID), nil
}

// ============================================================
// JSON RESPONSE
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

	_ = json.NewEncoder(w).Encode(data)
}

// ============================================================
// ERROR RESPONSE
// ============================================================

func writeError(
	w http.ResponseWriter,
	status int,
	message string,
) {
	writeJSON(
		w,
		status,
		map[string]interface{}{
			"success": false,
			"message": message,
		},
	)
}
