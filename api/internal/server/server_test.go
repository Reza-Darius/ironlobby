package server

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/reza-darius/ironlobby/internal/database"
	"github.com/reza-darius/ironlobby/internal/utils"
	"github.com/stretchr/testify/assert"
)

var testApp Application

func TestMain(m *testing.M) {
	utils.InitLogging()
	ctx := context.Background()

	db, pg := database.NewTestDBContainer()

	testApp = Application{
		db: db,
	}

	code := m.Run()

	if err := pg.Terminate(ctx); err != nil {
		log.Printf("failed to terminate container: %v", err)
	}
	log.Print("shut down container")

	os.Exit(code)
}

func TestHealthHandler(t *testing.T) {
	srv := utils.NewTestServer(t, testApp.routes())
	defer srv.Close()

	res, err := srv.Client().Get(srv.URL + "/api/health")
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)

	rb, err := io.ReadAll(res.Body)
	defer res.Body.Close()
	if err != nil {
		t.Fatalf("failed to read response body, err: %v", err)
	}

	t.Logf("response body: %v", string(rb))
}

func TestCreateUser(t *testing.T) {
	srv := utils.NewTestServer(t, testApp.routes())
	defer srv.Close()

	username := struct {
		Username string
	}{
		Username: "PrayDemon",
	}
	out, err := json.Marshal(username)
	if err != nil {
		t.Fatalf("failed to marshal username")
	}

	res, err := srv.Client().Post(srv.URL+"/api/user", "application/json", bytes.NewBuffer(out))
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode)

	url, err := url.Parse(srv.URL + "/api/")
	if err != nil {
		t.Fatalf("failed to parse url, err: %v", err)
	}
	cookie := srv.Client().Jar.Cookies(url)

	assert.Equal(t, 1, len(cookie))

	id, err := uuid.Parse(cookie[0].Value)
	if err != nil {
		t.Fatalf("failed to parse uuid from cookie, err: %v", err)
	}

	name, err := testApp.db.GetUser(t.Context(), id)
	if err != nil {
		t.Fatalf("failed to fetch user from db, err: %v", err)
	}

	assert.Equal(t, name, username.Username, "expecting user name from input and db to be the same")

	res, err = srv.Client().Post(srv.URL+"/api/user", "application/json", bytes.NewBuffer(out))
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	assert.Equal(t, http.StatusConflict, res.StatusCode, "the server should reject duplicate names")
}

func TestCreateLobby(t *testing.T) {
	srv := utils.NewTestServer(t, testApp.routes())
	defer srv.Close()

	// try to post lobby without a registered user
	res, err := srv.Client().Post(srv.URL+"/api/lobby", "application/json", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "we expect unauthorized")

	username := struct {
		Username string
	}{
		Username: "Skrt",
	}
	out, err := json.Marshal(username)
	if err != nil {
		t.Fatalf("failed to marshal username")
	}

	// register new user
	res, err = srv.Client().Post(srv.URL+"/api/user", "application/json", bytes.NewBuffer(out))
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode, "we should be able to create a user")

	// Golang stores time in nanoseconds, and postgres in microseconds
	// this truncation is only necessary for testing
	startDate := time.Now().AddDate(0, 0, 7).Truncate(time.Microsecond)
	lobbyCreateBody := database.InsertLobbyParams{
		LobbyName:   "Historical PVP",
		StartsAt:    startDate,
		Gamemode:    database.GamemodeVanilla,
		Description: "schizo lobby",
	}

	out, err = json.Marshal(lobbyCreateBody)
	if err != nil {
		t.Fatalf("failed to marshal username")
	}

	// create new lobby
	res, err = srv.Client().Post(srv.URL+"/api/lobby", "application/json", bytes.NewBuffer(out))
	defer res.Body.Close()
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode, "we should be able to create a lobby")

	var lobbyID int64
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read resp body")
	}

	err = json.Unmarshal(resBody, &lobbyID)
	if err != nil {
		t.Fatalf("failed to unmarshal id")
	}

	// fetch newly created lobby
	res, err = srv.Client().Get(srv.URL + "/api/lobby/" + strconv.FormatInt(lobbyID, 10))
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode, "we should be able to query the lobby after creating it")

	var lobby database.LobbyInfo

	resBody, err = io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read resp body")
	}

	err = json.Unmarshal(resBody, &lobby)
	if err != nil {
		t.Fatalf("failed to unmarshal id")
	}

	assert.Equal(t, lobby.Lobby.Description, lobbyCreateBody.Description, "description should match")
	assert.Equal(t, lobby.Lobby.Gamemode, lobbyCreateBody.Gamemode, "gamemode should match")
	assert.Equal(t, lobby.Lobby.LobbyName, lobbyCreateBody.LobbyName, "lobby name should match")
	assert.Equal(t, lobby.Lobby.StartsAt, startDate, "start date should match")
}

func TestJoinLobby(t *testing.T) {
	srv := utils.NewTestServer(t, testApp.routes())
	defer srv.Close()

	username := struct {
		Username string
	}{
		Username: "Inno",
	}
	out, err := json.Marshal(username)
	if err != nil {
		t.Fatalf("failed to marshal username")
	}

	// register new user
	res, err := srv.Client().Post(srv.URL+"/api/user", "application/json", bytes.NewBuffer(out))
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode, "we should be able to create a user")

	// Golang stores time in nanoseconds, and postgres in microseconds
	// this truncation is only necessary for testing
	startDate := time.Now().AddDate(0, 0, 7).Truncate(time.Microsecond)
	lobbyCreateBody := database.InsertLobbyParams{
		LobbyName:   "Historical PVP",
		StartsAt:    startDate,
		Gamemode:    database.GamemodeVanilla,
		Description: "schizo lobby",
	}

	out, err = json.Marshal(lobbyCreateBody)
	if err != nil {
		t.Fatalf("failed to marshal username")
	}

	// create new lobby
	res, err = srv.Client().Post(srv.URL+"/api/lobby", "application/json", bytes.NewBuffer(out))
	defer res.Body.Close()
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	assert.Equal(t, http.StatusOK, res.StatusCode, "we should be able to create a lobby")

	var lobbyID int64
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		t.Fatalf("failed to read resp body")
	}

	err = json.Unmarshal(resBody, &lobbyID)
	if err != nil {
		t.Fatalf("failed to unmarshal id")
	}

	// join lobby
	joinParam := database.AssignPlayerToLobbyParams{
		CountryTag: "GER",
	}

	out, err = json.Marshal(joinParam)
	if err != nil {
		t.Fatalf("failed to marshal username")
	}

	url := srv.URL + "/api/lobby/" + strconv.Itoa(int(lobbyID)) + "/player"
	res, err = srv.Client().Post(url, "application/json", bytes.NewBuffer(out))
	defer res.Body.Close()
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}
}
