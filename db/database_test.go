package db_test

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v4"
	"github.com/pimka/go-onenote/db"
	"testing"
	"time"
)

const DBURL = "postgres://puser:puser123@localhost:5432/postgres?sslmode=disable"

type MockDB struct {
	Notes []*db.Note
}

func (m *MockDB) Create(ctx context.Context, uid uuid.UUID, text string, exp_time int) (*db.Note, error) {
	note := m.push(text, exp_time)
	return note, nil
}

func (m *MockDB) Get(ctx context.Context, uid uuid.UUID) (*db.Note, error) {
	for _, n := range m.Notes {
		if n.ID == uid {
			return n, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *MockDB) List(ctx context.Context) ([]*db.Note, error) {
	return m.Notes, nil
}

func (m *MockDB) Update(ctx context.Context, uid uuid.UUID, newText string) (*db.Note, error) {
	for _, n := range m.Notes {
		if n.ID == uid {
			n.Text = newText
			return n, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *MockDB) Delete(ctx context.Context, uid uuid.UUID) (*db.Note, error) {
	delIdx := -1
	var note *db.Note
	for idx, n := range m.Notes {
		if n.ID == uid {
			note = n
			delIdx = idx
		}
	}
	if delIdx == -1 {
		return nil, pgx.ErrNoRows
	}

	notes := m.Notes
	notes[len(notes)-1], notes[delIdx] = notes[delIdx], notes[len(notes)-1]
	m.Notes = notes
	return note, nil
}

func (m *MockDB) ClearExpired(ctx context.Context) error {
	var newNotes []*db.Note
	now := time.Now()
	for _, n := range m.Notes {
		if now.After(n.Created.Add(time.Minute * time.Duration(n.Expiration))) {
			newNotes = append(newNotes, n)
		}
	}
	m.Notes = newNotes
	return nil
}

func NewMockDB() db.NoteHandler {
	mdb := &MockDB{}
	for i := 0; i < 20; i++ {
		mdb.push(fmt.Sprintf("pupa-test-%d", i), i)
	}
	return mdb
}

func (m *MockDB) push(text string, expiration int) *db.Note {
	created := time.Now()
	uid, _ := uuid.NewV4()
	note := &db.Note{
		ID:         uid,
		Text:       text,
		Created:    created,
		Expiration: expiration,
	}
	m.Notes = append(m.Notes, note)
	return note
}

func TestNoteDB(t *testing.T) {
	ndb := NewMockDB()

	ctx := context.Background()
	uid, err := uuid.NewV4()
	note, err := ndb.Create(ctx, uid, "test message", 10)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(note)

	err = ndb.ClearExpired(ctx)
	if err != nil {
		t.Fatal(err)
	}

	n, err := ndb.Get(ctx, uid)
	if err != nil {
		t.Fatal(err)
	}
	if uid != n.ID {
		t.Fatal("GET note.ID != uid")
	}

	notes, err := ndb.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(notes)

	newNote, err := ndb.Update(ctx, uid, "memes-pepes")
	if err != nil {
		t.Fatal(err)
	}
	if newNote.Text != note.Text && newNote.ID == note.ID && note.Created.Equal(newNote.Created) {
		t.Fatal("That's another note")
	}

	for _, n := range notes {
		note, err = ndb.Delete(ctx, n.ID)
		if err != nil {
			t.Fatal(err)
		}
		if n.ID != note.ID {
			t.Fatal("DELETE note.ID != uid")
		}
	}
}
