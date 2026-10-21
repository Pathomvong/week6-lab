package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// UpsertReview สร้างรีวิวใหม่หรือแก้รีวิวเดิมของสมาชิกคนเดิม
func (s *Store) UpsertReview(ctx context.Context, memberID, movieID int64, text string) (Review, error) {
	var r Review

	err := s.Pool.QueryRow(ctx, `
		WITH saved AS (
			INSERT INTO reviews (member_id, movie_id, text)
			VALUES ($1, $2, $3)
			ON CONFLICT (member_id, movie_id)
			DO UPDATE SET text = EXCLUDED.text, updated_at = now()
			RETURNING id, member_id, movie_id, text, created_at, updated_at
		)
		SELECT saved.id, saved.member_id, saved.movie_id, saved.text,
		       saved.created_at, saved.updated_at, m.display_name, v.score
		FROM saved
		JOIN members m ON m.id = saved.member_id
		LEFT JOIN votes v
		  ON v.member_id = saved.member_id AND v.movie_id = saved.movie_id
	`, memberID, movieID, text).Scan(
		&r.ID, &r.MemberID, &r.MovieID, &r.Text,
		&r.CreatedAt, &r.UpdatedAt, &r.Member.DisplayName, &r.Score,
	)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return Review{}, ErrNotFound
	}
	if err != nil {
		return Review{}, err
	}

	r.Member.ID = r.MemberID
	return r, nil
}
