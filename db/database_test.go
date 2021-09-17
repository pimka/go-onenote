package db_test

import (
	"context"
	"github.com/gofrs/uuid"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/pimka/go-onenote/db"
	"testing"
)

const DBURL = "postgres://puser:puser123@localhost:5432/postgres?sslmode=disable"

func TestNoteDB(t *testing.T) {
	ndb := db.NewMockDB()

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
