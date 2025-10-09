package sent

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SentItem struct {
	URL string
}

type SentDB struct {
	pool  *pgxpool.Pool
	items []SentItem
}

func NewSentDB() (*SentDB, error) {
	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}
	return &SentDB{pool: dbpool}, nil
}

func (instance *SentDB) HasBeenSent(url string) (bool, error) {
	if instance.items == nil {
		qitems, err := instance.querySentItems()
		if err != nil {
			return false, err
		}
		instance.items = qitems
	}
	for _, item := range instance.items {
		if item.URL == url {
			return true, nil
		}
	}
	return false, nil

}

func (instance *SentDB) MarkAsSent(url string) error {
	// Begin a transaction
	tx, err := instance.pool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}
	defer tx.Rollback(context.Background())

	_, err = tx.Exec(context.Background(), "INSERT INTO sent_items (url, ts) VALUES ($1, $2)", url, time.Now().UnixMilli())
	if err != nil {
		return fmt.Errorf("error marking item as sent: %v", err)
	}
	instance.items = nil // reset cache
	return tx.Commit(context.Background())
}

func (instance *SentDB) querySentItems() ([]SentItem, error) {
	rows, err := instance.pool.Query(context.Background(), "SELECT TOP 50 url FROM sent_items order by ts DESC")
	if err != nil {
		return nil, fmt.Errorf("error querying sent items: %v", err)
	}
	defer rows.Close()

	var items []SentItem
	for rows.Next() {
		var item SentItem
		if err := rows.Scan(&item.URL); err != nil {
			return nil, fmt.Errorf("error scanning sent item: %v", err)
		}
		items = append(items, item)
	}
	return items, nil
}
