package dao

import (
	"testing"

	db "github.com/wweitzel/top90/internal/db/models"
	"gotest.tools/v3/assert"
)

func TestPlayersDao(t *testing.T) {
	t.Parallel()

	dao, pool, resource, err := createTestDb()
	assert.NilError(t, err)
	defer pool.Purge(resource)

	player := db.Player{
		Id:          1,
		Name:        "jim bob",
		FirstName:   "jim",
		LastName:    "bob",
		Age:         21,
		Nationality: "usa",
		Height:      "100 cm",
		Weight:      "100 kg",
		Photo:       "https://somephoto",
	}
	insertedPlayer, err := dao.UpsertPlayer(player)
	assert.NilError(t, err)
	player.CreatedAt = insertedPlayer.CreatedAt
	assert.DeepEqual(t, insertedPlayer, player)

	player = db.Player{
		Id:          1,
		Name:        "jimmy bob",
		FirstName:   "jimmy",
		LastName:    "bob bob",
		Age:         30,
		Nationality: "united states",
		Height:      "200 cm",
		Weight:      "200 kg",
		Photo:       "https://somephoto2",
	}
	insertedPlayer, err = dao.UpsertPlayer(player)
	player.CreatedAt = insertedPlayer.CreatedAt
	assert.NilError(t, err)
	assert.DeepEqual(t, insertedPlayer, player)
}
