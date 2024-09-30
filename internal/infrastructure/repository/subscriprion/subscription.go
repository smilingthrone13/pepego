package subscriprion

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

func (r *Repository) Get(ctx context.Context, chatId string) (sub domain.Subscription, err error) {
	query := "SELECT chat_id, created_at, period FROM subscription WHERE chat_id = ?"
	err = r.db.Conn().QueryRowContext(ctx, query, chatId).Scan(&sub.ChatId, &sub.CreatedAt, &sub.Period)
	if err != nil {
		return sub, errors.Wrap(err, "can not get subscription")
	}

	return sub, nil
}

func (r *Repository) GetAll(ctx context.Context) (subs []domain.Subscription, err error) {
	query := "SELECT chat_id, created_at, period FROM subscription"
	rows, err := r.db.Conn().QueryContext(ctx, query)
	if err != nil {
		return nil, errors.Wrap(err, "can not exec query")
	}
	defer rows.Close()

	for rows.Next() {
		var sub domain.Subscription

		if err = rows.Scan(&sub.ChatId, &sub.CreatedAt, &sub.Period); err != nil {
			return nil, errors.Wrap(err, "can not scan row")
		}

		subs = append(subs, sub)
	}

	if err = rows.Err(); err != nil {
		return nil, errors.Wrap(err, "can not read rows")
	}

	return subs, nil
}

func (r *Repository) Create(ctx context.Context, sub domain.Subscription) error {
	query := `
	INSERT INTO subscription (chat_id, created_at, period)
	VALUES (?, ?, ?)
	ON CONFLICT(chat_id) DO UPDATE SET created_at=excluded.created_at, period=excluded.period
	`
	_, err := r.db.Conn().ExecContext(ctx, query, sub.ChatId, sub.CreatedAt, sub.Period)
	if err != nil {
		return errors.Wrap(err, "can not exec query")
	}

	return nil
}

func (r *Repository) Delete(ctx context.Context, chatId string) error {
	query := "DELETE FROM subscription WHERE chat_id = ?"
	_, err := r.db.Conn().ExecContext(ctx, query, chatId)
	if err != nil {
		return errors.Wrap(err, "can not exec query")
	}

	return nil
}
