package database

import (
	"context"
	"log"
	"os"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/reza-darius/ironlobby/internal/utils"
	"github.com/stretchr/testify/assert"
)


var testDB *Queries

func TestMain(m *testing.M) {
	utils.InitLogging()
	ctx := context.Background()

	db, pg := NewTestDBContainer()

	testDB = New(db.pool)

	code := m.Run()

	if err := pg.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}
	log.Print("shut down container")

	os.Exit(code)
}

func TestInsert(t *testing.T) {
	ctx := context.Background()
	playerNames := []string{
		"skrt",
		"inno",
		"pray",
	}

	regPlayer := make(map[string]uuid.UUID)

	var hostID uuid.UUID
	for _, name := range playerNames {
		player, err := testDB.InsertNewPlayer(ctx, name)
		if err != nil {
			t.Fatalf("failed to insert player %v", err)
		}
		t.Logf("inserterd player %v, id = %v", player.PlayerName, player.ID)
		hostID = player.ID
		regPlayer[name] = player.ID
	}

	allPlayers, err := testDB.GetAllPlayer(ctx)
	if err != nil {
		t.Fatalf("failed to fetch players err = %v", err)
	}
	assert.Equal(t, len(allPlayers), 3)

	lobby, err := testDB.InsertLobby(ctx, InsertLobbyParams{
		HostPlayer: hostID,
		LobbyName:  "test lobby",
		StartsAt:   time.Now().AddDate(0, 0, 7),
		Gamemode:   GamemodeVanilla,
	})
	if err != nil {
		t.Fatalf("failed to create lobby %v", err)
	}
	t.Logf("lobby created: %v", lobby)

	err = testDB.AddLobbyCountry(ctx, AddLobbyCountryParams{
		LobbyID:    lobby.ID,
		CountryTag: "GER",
		MaxSlots:   2,
	})
	if err != nil {
		t.Fatalf("failed to add lobby country GER %v", err)
	}
	err = testDB.AddLobbyCountry(ctx, AddLobbyCountryParams{
		LobbyID:    lobby.ID,
		CountryTag: "SOV",
		MaxSlots:   1,
	})
	if err != nil {
		t.Fatalf("failed to add lobby country SOV %v", err)
	}
	err = testDB.AddLobbyCountry(ctx, AddLobbyCountryParams{
		LobbyID:    lobby.ID,
		CountryTag: "JAP",
		MaxSlots:   1,
	})
	if err != nil {
		t.Fatalf("failed to add lobby country JAP %v", err)
	}

	lobbyInfo, err := testDB.GetLobbyCountries(ctx, lobby.ID)
	if err != nil {
		t.Fatalf("failed to fetch lobby countries %v", err)
	}

	assert.Equal(t, len(lobbyInfo), 3)

	for _, row := range lobbyInfo {
		t.Logf("lobby country row: tag = %v, occupied_slots = %v, max_slots = %v", row.CountryTag, row.OccupiedSlots, row.MaxSlots)
	}

	err = testDB.AssignPlayerToLobby(ctx, AssignPlayerToLobbyParams{
		LobbyID:    lobby.ID,
		CountryTag: "GER",
		PlayerID:   regPlayer["pray"],
	})
	if err != nil {
		t.Fatalf("failed to register pray to GER %v", err)
	}
	err = testDB.AssignPlayerToLobby(ctx, AssignPlayerToLobbyParams{
		LobbyID:    lobby.ID,
		CountryTag: "JAP",
		PlayerID:   regPlayer["inno"],
	})
	if err != nil {
		t.Fatalf("failed to register inno to JAP %v", err)
	}
	err = testDB.AssignPlayerToLobby(ctx, AssignPlayerToLobbyParams{
		LobbyID:    lobby.ID,
		CountryTag: "SOV",
		PlayerID:   regPlayer["skrt"],
	})
	if err != nil {
		t.Fatalf("failed to register skrt to SOV %v", err)
	}

	lobbyPlayers, err := testDB.GetLobbyPlayers(ctx, lobby.ID)
	if err != nil {
		t.Fatalf("failed to fetch lobby players %v", err)
	}
	assert.Equal(t, len(lobbyPlayers), 3)
	for _, row := range lobbyPlayers {
		t.Logf("%v", row)
	}
}
