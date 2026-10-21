package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strings"
	"unicode/utf8"

	"moviehub/mw"
	"moviehub/store"
)

type reviewWriteReq struct {
	Text string `json:"text"`
}

func (h *Handler) CreateReview(w http.ResponseWriter, r *http.Request) {
	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}

	movieID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}

	var req reviewWriteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}

	req.Text = strings.TrimSpace(req.Text)
	if utf8.RuneCountInString(req.Text) < 10 {
		writeError(w, http.StatusBadRequest, "BAD_REVIEW", "รีวิวต้องมีอย่างน้อย 10 ตัวอักษร")
		return
	}

	review, err := h.Store.UpsertReview(r.Context(), memberID, movieID, req.Text)
	if errors.Is(err, store.ErrNotFound) {
		writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
		return
	}
	if err != nil {
		internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, review)
}

type voteWriteReq struct {
	Score int `json:"score"`
}

func (h *Handler) PutVote(w http.ResponseWriter, r *http.Request) {
	memberID, ok := mw.MemberID(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "NO_TOKEN", "กรุณาเข้าสู่ระบบก่อน")
		return
	}

	movieID, ok := pathID(r, "id")
	if !ok {
		writeError(w, http.StatusBadRequest, "BAD_ID", "เลขหนังต้องเป็นตัวเลข")
		return
	}

	var req voteWriteReq
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "BAD_JSON", "อ่านกล่อง JSON ไม่ออก")
		return
	}

	if req.Score < 1 || req.Score > 10 {
		writeError(w, http.StatusBadRequest, "BAD_SCORE", "คะแนนต้องอยู่ระหว่าง 1 ถึง 10")
		return
	}

	if err := h.Store.UpsertVote(r.Context(), memberID, movieID, req.Score); err != nil {
		if errors.Is(err, store.ErrNotFound) {
			writeError(w, http.StatusNotFound, "MOVIE_NOT_FOUND", "ไม่พบหนังเรื่องนี้")
			return
		}
		internal(w, err)
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"movieId": movieID,
		"score":   req.Score,
	})
}
