package store

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
)

// ListWishlist คืนรายชื่อหนังใน Wishlist ของสมาชิกคนที่ login อยู่
func (s *Store) ListWishlist(ctx context.Context, memberID int64) ([]Movie, error) {
	rows, err := s.Pool.Query(ctx, `
		SELECT m.id, m.title, m.title_th,
			(SELECT coalesce(g.name_th, g.name_en)
			 FROM movie_genres mg
			 JOIN genres g ON g.id = mg.genre_id
			 WHERE mg.movie_id = m.id AND mg.position = 0),
			extract(year FROM m.release_date)::int,
			round(m.vote_average, 1)::float8,
			CASE
				WHEN m.poster_path LIKE 'http%' THEN m.poster_path
				ELSE 'https://image.tmdb.org/t/p/w500' || m.poster_path
			END
		FROM wishlist_items wi
		JOIN movies m ON m.id = wi.movie_id
		WHERE wi.member_id = $1
		ORDER BY wi.added_at DESC, m.id DESC
	`, memberID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	items := []Movie{}
	for rows.Next() {
		var m Movie
		if err := rows.Scan(
			&m.ID, &m.Title, &m.TitleTh, &m.Genre,
			&m.Year, &m.Rating, &m.Poster,
		); err != nil {
			return nil, err
		}
		items = append(items, m)
	}
	return items, rows.Err()
}

// AddWishlistItem เพิ่มหนัง โดยไม่เพิ่มซ้ำถ้าสมาชิกบันทึกไว้แล้ว
func (s *Store) AddWishlistItem(ctx context.Context, memberID, movieID int64) error {
	_, err := s.Pool.Exec(ctx, `
		INSERT INTO wishlist_items (member_id, movie_id)
		VALUES ($1, $2)
		ON CONFLICT (member_id, movie_id) DO NOTHING
	`, memberID, movieID)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23503" {
		return ErrNotFound
	}
	return err
}

// RemoveWishlistItem นำหนังออกจาก Wishlist ของสมาชิกคนนั้นเท่านั้น
func (s *Store) RemoveWishlistItem(ctx context.Context, memberID, movieID int64) error {
	_, err := s.Pool.Exec(ctx, `
		DELETE FROM wishlist_items
		WHERE member_id = $1 AND movie_id = $2
	`, memberID, movieID)
	return err
}
