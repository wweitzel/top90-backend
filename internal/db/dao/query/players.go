package query

import db "github.com/wweitzel/top90/internal/db/models"

func GetPlayer(id int) (string, []any) {
	query := "SELECT * FROM players WHERE id = $1"
	return query, []any{id}
}

func PlayerExists(id int) (string, []any) {
	query := "SELECT count(*) FROM players WHERE id = $1"
	var args []any
	args = append(args, id)
	return query, args
}

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

func SearchPlayers(searchTerm string) (string, []any) {
	query := `
		SELECT * FROM players
		WHERE SIMILARITY(first_name, $1) > 0.4 OR
		SIMILARITY(last_name, $2) > 0.4 OR
		SIMILARITY(CONCAT(first_name, ' ', last_name), $3) > 0.4 limit 10`
	var args []any
	args = append(args, searchTerm, searchTerm, searchTerm)
	return query, args
}
