package server_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/pimka/go-onenote/db"
	"github.com/pimka/go-onenote/server"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func createServer(mdb *db.MockDB) *server.Server {
	s := &server.Server{
		DBPurger: db.NewPurger(mdb, time.Minute, 5),
		Router:   nil,
		NH:       mdb,
	}
	return s
}

func TestServer_ListNotes(t *testing.T) {
	mdb := db.NewMockDB()
	s := createServer(mdb)

	req, err := http.NewRequest("GET", "/note/", nil)
	if err != nil {
		t.Fatal(err)
	}
	respRecoder := httptest.NewRecorder()
	s.ListNotes().ServeHTTP(respRecoder, req)
	if respRecoder.Code != http.StatusAccepted {
		t.Error("Server error on ListNotes")
	}
}

func TestServer_AddNote(t *testing.T) {
	mbd := db.NewMockDB()
	s := createServer(mbd)
	note := db.Note{
		Text:       "test",
		Expiration: 100,
	}
	noteJson, err := json.Marshal(note)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("POST", "/note/", bytes.NewBuffer(noteJson))
	if err != nil {
		t.Fatal(err)
	}
	respRecoder := httptest.NewRecorder()
	s.AddNote().ServeHTTP(respRecoder, req)
	if respRecoder.Code != http.StatusAccepted {
		t.Error("Server error on AddNote")
	}
}

func TestServer_GetNote(t *testing.T) {
	mbd := db.NewMockDB()
	s := createServer(mbd)
	uid := mbd.Notes[0].ID
	req, err := http.NewRequest("GET", fmt.Sprintf("/note/%s", uid), nil)
	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{
		"uid": uid.String(),
	})
	respRecoder := httptest.NewRecorder()
	s.GetNote().ServeHTTP(respRecoder, req)
	if respRecoder.Code != http.StatusOK {
		t.Error("Server error on GetNote")
	}
}

func TestServer_UpdateNote(t *testing.T) {
	mbd := db.NewMockDB()
	s := createServer(mbd)
	note := mbd.Notes[0]
	newNote := db.Note{
		Text: "new test",
	}
	noteJson, err := json.Marshal(newNote)
	if err != nil {
		t.Fatal(err)
	}
	req, err := http.NewRequest("PATCH", fmt.Sprintf("/note/%s", note.ID), bytes.NewBuffer(noteJson))
	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{
		"uid": note.ID.String(),
	})
	respRecoder := httptest.NewRecorder()
	s.UpdateNote().ServeHTTP(respRecoder, req)
	if respRecoder.Code != http.StatusAccepted {
		t.Error("Server error on UpdateNote")
	}
	if mbd.Notes[0].Text != newNote.Text {
		t.Error("Doesnt updated")
	}
}

func TestServer_DeleteNote(t *testing.T) {
	mbd := db.NewMockDB()
	s := createServer(mbd)
	uid := mbd.Notes[0].ID
	req, err := http.NewRequest("DELETE", fmt.Sprintf("/note/%s", uid), nil)
	if err != nil {
		t.Fatal(err)
	}

	req = mux.SetURLVars(req, map[string]string{
		"uid": uid.String(),
	})
	respRecoder := httptest.NewRecorder()
	s.DeleteNote().ServeHTTP(respRecoder, req)
	if respRecoder.Code != http.StatusNoContent {
		t.Error("Server error on DeleteNote")
	}
	if uid == mbd.Notes[0].ID {
		t.Error("Doesnt deleted")
	}
}

func TestServer_PopNote(t *testing.T) {
	mbd := db.NewMockDB()
	s := createServer(mbd)
	uid := mbd.Notes[0].ID
	note := db.Note{
		ID: uid,
	}
	noteJson, err := json.Marshal(note)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("DELETE", "/note/api", bytes.NewBuffer(noteJson))
	if err != nil {
		t.Fatal(err)
	}

	respRecoder := httptest.NewRecorder()
	s.PopNote().ServeHTTP(respRecoder, req)
	if respRecoder.Code != http.StatusOK {
		t.Error("Server error on DeleteNote")
	}
	if uid == mbd.Notes[0].ID {
		t.Error("Doesnt pop")
	}
}

func TestServer_PeekNote(t *testing.T) {
	mbd := db.NewMockDB()
	s := createServer(mbd)
	uid := mbd.Notes[0].ID
	note := db.Note{
		ID: uid,
	}
	noteJson, err := json.Marshal(note)
	if err != nil {
		t.Fatal(err)
	}

	req, err := http.NewRequest("GET", "/note/api", bytes.NewBuffer(noteJson))
	if err != nil {
		t.Fatal(err)
	}

	respRecoder := httptest.NewRecorder()
	s.PeekNote().ServeHTTP(respRecoder, req)
	if respRecoder.Code != http.StatusOK {
		t.Error("Server error on DeleteNote")
	}
}
