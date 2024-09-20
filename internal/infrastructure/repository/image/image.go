package image

import (
	"apubot/internal/domain"
	"apubot/internal/infrastructure/database"
	"context"
	"github.com/pkg/errors"
)

type Repository struct {
	db *database.DB
}

func New(db *database.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetAll(ctx context.Context) (map[string]string, error) {
	query := "SELECT name, tg_id FROM images"
	rows, err := r.db.Conn().QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "can not exec query")
	}
	defer rows.Close()

	images := make(map[string]string)
	for rows.Next() {
		var name, tgID string
		if err = rows.Scan(&name, &tgID); err != nil {
			return nil, errors.Wrap(err, "can not scan row")
		}
		images[name] = tgID
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "can not read rows")
	}

	return images, nil
}

func (r *Repository) SaveImage(ctx context.Context, file domain.File) error {
	query := "INSERT INTO images (name, tg_id) VALUES (?, ?) ON CONFLICT(name) DO UPDATE SET tg_id=excluded.tg_id;"
	_, err := r.db.Conn().ExecContext(ctx, query, file.Name, file.TgID)
	if err != nil {
		return errors.Wrap(err, "can not exec query")
	}

	return nil
}
