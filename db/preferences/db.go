package preferences

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type PreferenceItem struct {
	ID         int64
	UserID     int64
	Preference string
}
type PreferenceItems []PreferenceItem

func (p PreferenceItems) String() string {
	var prefs []string
	for _, item := range p {
		prefs = append(prefs, item.Preference)
	}
	return strings.Join(prefs, " ")
}

func GetPreferences(userID int64) (PreferenceItems, error) {

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return nil, fmt.Errorf("error connecting to database: %v", err)
	}
	defer dbpool.Close()

	rows, err := dbpool.Query(context.Background(), "SELECT id, user_id, preference FROM preferences WHERE user_id = $1", userID)
	if err != nil {
		return nil, fmt.Errorf("error querying preferences: %v", err)
	}
	defer rows.Close()

	var preferences []PreferenceItem
	for rows.Next() {
		var p PreferenceItem
		if err := rows.Scan(&p.ID, &p.UserID, &p.Preference); err != nil {
			fmt.Printf("error scanning preference: %v", err)
		}
		preferences = append(preferences, p)
	}
	return preferences, nil
}

func SetPreferences(userID int64, preferences []string) error {

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer dbpool.Close()

	// Begin a transaction
	tx, err := dbpool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}

	// Delete existing preferences
	_, err = tx.Exec(context.Background(), "DELETE FROM preferences WHERE user_id = $1", userID)
	if err != nil {
		tx.Rollback(context.Background())
		return fmt.Errorf("error deleting preferences: %v", err)
	}

	// Insert new preferences
	for _, pref := range preferences {
		_, err = tx.Exec(context.Background(), "INSERT INTO preferences (user_id, preference) VALUES ($1, $2)", userID, pref)
		if err != nil {
			tx.Rollback(context.Background())
			return fmt.Errorf("error saving preference: %v", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}

func AddPreferences(userID int64, preferences []string) error {

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer dbpool.Close()

	// Begin a transaction
	tx, err := dbpool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}

	// Insert new preferences
	for _, pref := range preferences {
		_, err = tx.Exec(context.Background(), "INSERT INTO preferences (user_id, preference) VALUES ($1, $2)", userID, pref)
		if err != nil {
			tx.Rollback(context.Background())
			return fmt.Errorf("error saving preference: %v", err)
		}
	}

	// Commit the transaction
	if err = tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}

func RemovePreferences(userID int64, preferences []string) error {

	dbpool, err := pgxpool.New(context.Background(), os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("error connecting to database: %v", err)
	}
	defer dbpool.Close()

	// Begin a transaction
	tx, err := dbpool.Begin(context.Background())
	if err != nil {
		return fmt.Errorf("error beginning transaction: %v", err)
	}

	// Delete preferences
	for _, pref := range preferences {
		a, err := tx.Exec(context.Background(), "DELETE FROM preferences WHERE user_id = $1 AND preference = $2", userID, pref)
		if err != nil {
			tx.Rollback(context.Background())
			return fmt.Errorf("error deleting preference: %v", err)
		}
		if a.RowsAffected() == 0 {
			tx.Rollback(context.Background())
			return fmt.Errorf("no rows deleted")
		}
	}

	// Commit the transaction
	if err = tx.Commit(context.Background()); err != nil {
		return fmt.Errorf("error committing transaction: %v", err)
	}
	return nil
}
