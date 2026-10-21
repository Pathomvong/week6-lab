package handler

import (
	"errors"
	"net/http"

	"moviehub/mw"
	"moviehub/store"
)

// GetWishlist แสดง Wishlist ของสมาชิกที่ login อยู่
func (h *Handler) GetWishlist(w http.ResponseWriter, r *http.Request) {
	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}

	items, err := h.Store.ListWishlist(r.Context(), memberID)
	if err != nil {
		internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{"items": items})
}

// AddWishlistItem เพิ่มหนังเข้ารายการที่อยากดู
func (h *Handler) AddWishlistItem(w http.ResponseWriter, r *http.Request) {
	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}

	movieID, ok := pathID(r, "movieId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}

	err := h.Store.AddWishlistItem(r.Context(), memberID, movieID)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

// RemoveWishlistItem นำหนังออกจากรายการที่อยากดู
func (h *Handler) RemoveWishlistItem(w http.ResponseWriter, r *http.Request) {
	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}

	movieID, ok := pathID(r, "movieId")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}

	if err := h.Store.RemoveWishlistItem(r.Context(), memberID, movieID); err != nil {
		internal(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
