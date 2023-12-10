package query

import db "github.com/wweitzel/top90/internal/db/models"

func UpsertPlayer(player db.Player) (string, []any) {
	query := `
		INSERT INTO players (id, name, first_name, last_name, age, nationality, height, weight, photo)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
		ON CONFLICT (id) DO UPDATE SET name = $10, first_name = $11, last_name = $12, age = $13, nationality = $14, height = $15, weight = $16, photo = $17
		RETURNING *`
	var args []any
	args = append(args,
		player.Id,
		player.Name,
		player.FirstName,
		player.LastName,
		player.Age,
		player.Nationality,
		player.Height,
		player.Weight,
		player.Photo,
		player.Name,
		player.FirstName,
		player.LastName,
		player.Age,
		player.Nationality,
		player.Height,
		player.Weight,
		player.Photo)
	return query, args
}
