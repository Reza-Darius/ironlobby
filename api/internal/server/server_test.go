package server

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
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
		config: &utils.AppConfig{
			DebugCors: true,
		},
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

func NewTestUser(srv *httptest.Server, name string) error {
	username := struct {
		Username string
	}{
		Username: name,
	}
	out, err := json.Marshal(username)
	if err != nil {
		return err
	}

	res, err := srv.Client().Post(srv.URL+"/api/user", "application/json", bytes.NewBuffer(out))
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to create user, status: %v", res.StatusCode)
	}

	return nil
}

func CreateTestLobby(srv *httptest.Server, lobby *database.InsertLobbyParams, countries []database.UpsertLobbyCountryParams) (int64, error) {
	// Golang stores time in nanoseconds, and postgres in microseconds
	// this truncation is only necessary for testing to test the output
	lobby.StartsAt = lobby.StartsAt.Truncate(time.Microsecond)

	params := database.CreateLobbyRequest{
		Lobby:     *lobby,
		Countries: countries,
	}
	out, err := json.Marshal(params)
	if err != nil {
		return 0, err
	}

	res, err := srv.Client().Post(srv.URL+"/api/lobby", "application/json", bytes.NewBuffer(out))
	defer res.Body.Close()
	if err != nil {
		return 0, err
	}

	if res.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("failed to create lobby, status: %v", res.StatusCode)
	}

	var lobbyID int64
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return 0, err
	}

	err = json.Unmarshal(resBody, &lobbyID)
	if err != nil {
		return 0, err
	}

	return lobbyID, nil
}

func AddTestCountry(srv *httptest.Server, args *database.UpsertLobbyCountryParams) error {
	addCountryJSON, err := json.Marshal(args)
	if err != nil {
		return err
	}

	url := srv.URL + "/api/lobby/" + strconv.Itoa(int(args.LobbyID)) + "/country"
	res, err := srv.Client().Post(url, "application/json", bytes.NewBuffer(addCountryJSON))
	defer res.Body.Close()
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusOK {
		return fmt.Errorf("couldnt add country, code: %v", res.StatusCode)
	}
	return nil
}

func TestCreateUser(t *testing.T) {
	srv := utils.NewTestServer(t, testApp.routes())
	defer srv.Close()

	username := "PrayDemon"
	err := NewTestUser(srv, username)
	if err != nil {
		t.Fatalf("failed to create user, err: %v", err)
	}

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

	assert.Equal(t, name, username, "expecting user name from input and db to be the same")
	assert.Error(t, NewTestUser(srv, username), "duplicate names should fail")
}

func TestCreateLobby(t *testing.T) {
	srv := utils.NewTestServer(t, testApp.routes())
	defer srv.Close()

	// try to post lobby without a registered user
	res, err := srv.Client().Post(srv.URL+"/api/lobby", "application/json", bytes.NewBuffer([]byte("")))
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	if !assert.Equal(t, http.StatusUnauthorized, res.StatusCode, "we expect unauthorized") {
		t.FailNow()
	}

	err = NewTestUser(srv, "Skrt")
	if err != nil {
		t.Fatalf("failed to create user, err: %v", err)
	}

	// Golang stores time in nanoseconds, and postgres in microseconds
	// this truncation is only necessary for testing
	lobbyCreateBody := database.InsertLobbyParams{
		LobbyName:   "Historical PVP",
		StartsAt:    time.Now().AddDate(0, 0, 7),
		Gamemode:    database.GamemodeVanilla,
		Description: "schizo lobby",
	}

	// create new lobby
	lobbyID, err := CreateTestLobby(srv, &lobbyCreateBody, nil)
	if err != nil {
		t.Fatalf("failed to create lobbs %v", err)
	}

	// fetch newly created lobby
	res, err = srv.Client().Get(srv.URL + "/api/lobby/" + strconv.FormatInt(lobbyID, 10))
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	if !assert.Equal(t, http.StatusOK, res.StatusCode, "we should be able to query the lobby after creating it") {
		t.FailNow()
	}

	var lobby database.LobbyInfo

	resBody, err := io.ReadAll(res.Body)
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
	assert.Equal(t, lobby.Lobby.StartsAt, lobbyCreateBody.StartsAt, "start date should match")
}

func TestCreateLobbyWithCountries(t *testing.T) {
	srv := utils.NewTestServer(t, testApp.routes())
	defer srv.Close()

	err := NewTestUser(srv, "player3")
	if err != nil {
		t.Fatalf("failed to create user, err: %v", err)
	}

	// Golang stores time in nanoseconds, and postgres in microseconds
	// this truncation is only necessary for testing
	lobbyCreateBody := database.InsertLobbyParams{
		LobbyName:   "Historical PVP",
		StartsAt:    time.Now().AddDate(0, 0, 7),
		Gamemode:    database.GamemodeVanilla,
		Description: "schizo lobby",
	}

	// create new lobby
	lobbyID, err := CreateTestLobby(srv, &lobbyCreateBody, []database.UpsertLobbyCountryParams{
		{
			CountryTag: "GER",
			MaxSlots:   1,
		},
		{
			CountryTag: "SOV",
			MaxSlots:   2,
		},
	})
	if err != nil {
		t.Fatalf("failed to create lobbs %v", err)
	}

	// fetch newly created lobby
	res, err := srv.Client().Get(srv.URL + "/api/lobby/" + strconv.FormatInt(lobbyID, 10))
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	if !assert.Equal(t, http.StatusOK, res.StatusCode, "we should be able to query the lobby after creating it") {
		t.FailNow()
	}

	var lobby database.LobbyInfo

	resBody, err := io.ReadAll(res.Body)
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
	assert.Equal(t, lobby.Lobby.StartsAt, lobbyCreateBody.StartsAt, "start date should match")
	assert.Equal(t, 2, len(lobby.Countries), "we added two countries")
}

func TestJoinLobby(t *testing.T) {
	srv := utils.NewTestServer(t, testApp.routes())
	defer srv.Close()

	username := "Inno"
	err := NewTestUser(srv, username)
	if err != nil {
		t.Fatalf("failed to create user, err: %v", err)
	}

	lobbyCreateBody := database.InsertLobbyParams{
		LobbyName:   "Historical PVP",
		StartsAt:    time.Now().AddDate(0, 0, 7),
		Gamemode:    database.GamemodeVanilla,
		Description: "schizo lobby",
	}

	lobbyID, err := CreateTestLobby(srv, &lobbyCreateBody, nil)
	if err != nil {
		t.Fatalf("failed to create lobbs %v", err)
	}

	// join lobby fail
	joinParam := database.UpsertPlayerLobbyParams{
		CountryTag: "GER",
	}

	joinLobbyJSON, err := json.Marshal(joinParam)
	if err != nil {
		t.Fatalf("failed to marshal username")
	}

	url := srv.URL + "/api/lobby/" + strconv.Itoa(int(lobbyID)) + "/player"
	res, err := srv.Client().Post(url, "application/json", bytes.NewBuffer(joinLobbyJSON))
	defer res.Body.Close()
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	if !assert.Equal(t, http.StatusBadRequest, res.StatusCode, "the country wasnt added yet") {
		t.FailNow()
	}

	// add countries GER and ITA
	addCountryParam := database.UpsertLobbyCountryParams{
		LobbyID:    lobbyID,
		CountryTag: "GER",
		MaxSlots:   1,
	}

	err = AddTestCountry(srv, &addCountryParam)
	if err != nil {
		t.Fatalf("failed to add GER, err: %v", err)
	}

	addCountryParam.CountryTag = "ITA"

	err = AddTestCountry(srv, &addCountryParam)
	if err != nil {
		t.Fatalf("failed to add ITA, err: %v", err)
	}

	// join lobby
	url = srv.URL + "/api/lobby/" + strconv.Itoa(int(lobbyID)) + "/player"
	res, err = srv.Client().Post(url, "application/json", bytes.NewBuffer(joinLobbyJSON))
	defer res.Body.Close()
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	if !assert.Equal(t, http.StatusOK, res.StatusCode, "we should be able to join as GER after adding it") {
		t.FailNow()
	}

	// check lobby
	lobby, err := srv.Client().Get(srv.URL + "/api/lobby/" + strconv.Itoa(int(lobbyID)))
	defer lobby.Body.Close()
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	body, err := io.ReadAll(lobby.Body)
	if err != nil {
		t.Fatalf("failed to read body, err: %v", err)
	}

	var lobbyParsed database.LobbyInfo
	err = json.Unmarshal(body, &lobbyParsed)
	if err != nil {
		t.Fatalf("failed to unmarshal body, err: %v", err)
	}

	assert.Equal(t, 2, len(lobbyParsed.Countries))
	assert.Equal(t, "GER", lobbyParsed.Countries[0].CountryTag)
	assert.Equal(t, int16(1), lobbyParsed.Countries[0].MaxSlots)
	assert.Equal(t, "ITA", lobbyParsed.Countries[1].CountryTag)
	assert.Equal(t, int16(1), lobbyParsed.Countries[1].MaxSlots)

	assert.Equal(t, 1, len(lobbyParsed.Players))
	assert.Equal(t, "GER", lobbyParsed.Players[0].CountryTag)
	assert.Equal(t, username, lobbyParsed.Players[0].PlayerName)
	assert.Equal(t, lobbyID, lobbyParsed.Players[0].LobbyID)

	// swap slots
	joinParam = database.UpsertPlayerLobbyParams{
		CountryTag: "ITA",
	}

	joinLobbyJSON, err = json.Marshal(joinParam)
	if err != nil {
		t.Fatalf("failed to marshal username")
	}

	url = srv.URL + "/api/lobby/" + strconv.Itoa(int(lobbyID)) + "/player"
	res, err = srv.Client().Post(url, "application/json", bytes.NewBuffer(joinLobbyJSON))
	defer res.Body.Close()
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	if !assert.Equal(t, http.StatusOK, res.StatusCode, "we should be able to swap") {
		t.FailNow()
	}

	// check lobby again
	lobby, err = srv.Client().Get(srv.URL + "/api/lobby/" + strconv.Itoa(int(lobbyID)))
	defer lobby.Body.Close()
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	body, err = io.ReadAll(lobby.Body)
	if err != nil {
		t.Fatalf("failed to read body, err: %v", err)
	}

	err = json.Unmarshal(body, &lobbyParsed)
	if err != nil {
		t.Fatalf("failed to unmarshal body, err: %v", err)
	}

	// country list should remain unchanged
	assert.Equal(t, 2, len(lobbyParsed.Countries))
	assert.Equal(t, "GER", lobbyParsed.Countries[0].CountryTag)
	assert.Equal(t, int16(1), lobbyParsed.Countries[0].MaxSlots)
	assert.Equal(t, "ITA", lobbyParsed.Countries[1].CountryTag)
	assert.Equal(t, int16(1), lobbyParsed.Countries[1].MaxSlots)

	// we are ITA now
	assert.Equal(t, 1, len(lobbyParsed.Players))
	assert.Equal(t, "ITA", lobbyParsed.Players[0].CountryTag)
	assert.Equal(t, username, lobbyParsed.Players[0].PlayerName)
	assert.Equal(t, lobbyID, lobbyParsed.Players[0].LobbyID)

	t.Logf("lobby: %s", body)
}

func TestUpdateLobby(t *testing.T) {
	srv := utils.NewTestServer(t, testApp.routes())
	defer srv.Close()

	err := NewTestUser(srv, "player1")
	if err != nil {
		t.Fatalf("failed to create user %v", err)
	}

	lobbyCreateBody := database.InsertLobbyParams{
		LobbyName:   "Historical PVP",
		StartsAt:    time.Now().AddDate(0, 0, 7),
		Gamemode:    database.GamemodeVanilla,
		Description: "schizo lobby",
	}

	lobbyID, err := CreateTestLobby(srv, &lobbyCreateBody, nil)
	if err != nil {
		t.Fatalf("failed to create lobby err: %v", err)
	}

	// update lobby description
	newLobbyName := "new lobby name"
	lobbyUpdateParams := database.UpdateLobbyParams{
		LobbyName: pgtype.Text{
			String: newLobbyName,
			Valid:  true,
		},
	}

	out, err := json.Marshal(&lobbyUpdateParams)
	if err != nil {
		t.Fatalf("failed to marshal new lobby params err: %v", err)
	}

	req, err := http.NewRequest("PATCH", srv.URL+"/api/lobby/"+strconv.Itoa(int(lobbyID)), bytes.NewBuffer(out))
	if err != nil {
		t.Fatalf("failed to create request err: %v", err)
	}

	res, err := srv.Client().Do(req)
	if err != nil {
		t.Fatalf("failed to send request err: %v", err)
	}

	if !assert.Equal(t, http.StatusOK, res.StatusCode, "update should work") {
		t.Fatalf("update didnt work")
	}

	// fetch update lobby
	lobby, err := srv.Client().Get(srv.URL + "/api/lobby/" + strconv.Itoa(int(lobbyID)))
	defer lobby.Body.Close()
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	body, err := io.ReadAll(lobby.Body)
	if err != nil {
		t.Fatalf("failed to read body, err: %v", err)
	}

	var lobbyParsed database.LobbyInfo
	err = json.Unmarshal(body, &lobbyParsed)
	if err != nil {
		t.Fatalf("failed to unmarshal body, err: %v", err)
	}

	if !assert.Equal(t, newLobbyName, lobbyParsed.Lobby.LobbyName, "update should work") {
		t.Fatalf("update didnt work")
	}
}

func TestDeleteLobby(t *testing.T) {
	srv := utils.NewTestServer(t, testApp.routes())
	defer srv.Close()

	err := NewTestUser(srv, "player2")
	if err != nil {
		t.Fatalf("failed to create user %v", err)
	}

	lobbyCreateBody := database.InsertLobbyParams{
		LobbyName:   "Historical PVP",
		StartsAt:    time.Now().AddDate(0, 0, 7),
		Gamemode:    database.GamemodeVanilla,
		Description: "schizo lobby",
	}

	lobbyID, err := CreateTestLobby(srv, &lobbyCreateBody, nil)
	if err != nil {
		t.Fatalf("failed to create lobby err: %v", err)
	}

	// add GER
	err = AddTestCountry(srv, &database.UpsertLobbyCountryParams{
		LobbyID:    lobbyID,
		CountryTag: "GER",
		MaxSlots:   1,
	})
	if err != nil {
		t.Fatalf("failed to add country err: %v", err)
	}

	// fetch lobby
	lobbyIDstr := strconv.Itoa(int(lobbyID))
	lobby, err := srv.Client().Get(srv.URL + "/api/lobby/" + lobbyIDstr)
	defer lobby.Body.Close()
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	body, err := io.ReadAll(lobby.Body)
	if err != nil {
		t.Fatalf("failed to read body, err: %v", err)
	}

	var lobbyParsed database.LobbyInfo
	err = json.Unmarshal(body, &lobbyParsed)
	if err != nil {
		t.Fatalf("failed to unmarshal body, err: %v", err)
	}

	if !assert.Equal(t, 1, len(lobbyParsed.Countries), "there should be one country") {
		t.Fatalf("adding country didnt work")
	}

	// delete GER
	req, err := http.NewRequest("DELETE", srv.URL+"/api/lobby/"+lobbyIDstr+"/country/GER", bytes.NewBuffer([]byte{}))
	if err != nil {
		t.Fatalf("failed to create request, err: %v", err)
	}
	_, err = srv.Client().Do(req)
	if err != nil {
		t.Fatalf("failed to send request, err: %v", err)
	}

	// check lobby
	lobby, err = srv.Client().Get(srv.URL + "/api/lobby/" + lobbyIDstr)
	if err != nil {
		t.Fatalf("failed to get a response, err: %v", err)
	}

	body, err = io.ReadAll(lobby.Body)
	if err != nil {
		t.Fatalf("failed to read body, err: %v", err)
	}

	err = json.Unmarshal(body, &lobbyParsed)
	if err != nil {
		t.Fatalf("failed to unmarshal body, err: %v", err)
	}

	if !assert.Equal(t, 0, len(lobbyParsed.Countries), "there should be no countries remaining") {
		t.Fatalf("delete didnt work")
	}
}
