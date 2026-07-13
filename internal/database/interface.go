package database

import (
	"context"

	"github.com/google/uuid"
)

func (db Database) OpenLobbies(ctx context.Context) (int64, error) {
	q := New(db.pool)
	return q.OpenLobbies(ctx)
}

func (db Database) NewUser(ctx context.Context, playerName string) (uuid.UUID, error) {
	q := New(db.pool)
	player, err := q.InsertNewPlayer(ctx, playerName)
	return player.ID, err
}
