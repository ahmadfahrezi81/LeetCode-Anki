package database

import (
	"database/sql"
)

func IncrementUserCoins(userID string, amount int) (int, error) {
	query := `
		UPDATE user_stats
		SET coins = coins + $1
		WHERE user_id = $2
		RETURNING coins
	`

	var newTotal int
	err := DB.QueryRow(query, amount, userID).Scan(&newTotal)
	if err == sql.ErrNoRows {
		// If user_stats doesn't exist, create it first
		_, err = CreateUserStats(userID)
		if err != nil {
			return 0, err
		}
		// Retry update
		err = DB.QueryRow(query, amount, userID).Scan(&newTotal)
	}

	return newTotal, err
}
