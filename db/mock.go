package db

import (
	"context"
	"fmt"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"time"
)

type MockDB struct {
	Notes []*Note
}

func (m *MockDB) Create(ctx context.Context, uid uuid.UUID, text string, exp_time int) (*Note, error) {
	note := m.push(text, exp_time)
	return note, nil
}

func (m *MockDB) Get(ctx context.Context, uid uuid.UUID) (*Note, error) {
	for _, n := range m.Notes {
		if n.ID == uid {
			return n, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *MockDB) List(ctx context.Context) ([]*Note, error) {
	return m.Notes, nil
}

func (m *MockDB) Update(ctx context.Context, uid uuid.UUID, newText string) (*Note, error) {
	for _, n := range m.Notes {
		if n.ID == uid {
			n.Text = newText
			return n, nil
		}
	}
	return nil, pgx.ErrNoRows
}

func (m *MockDB) Delete(ctx context.Context, uid uuid.UUID) (*Note, error) {
	delIdx := -1
	var note *Note
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
	var newNotes []*Note
	now := time.Now()
	for _, n := range m.Notes {
		if now.After(n.Created.Add(time.Minute * time.Duration(n.Expiration))) {
			newNotes = append(newNotes, n)
		}
	}
	m.Notes = newNotes
	return nil
}

func NewMockDB() *MockDB {
	mdb := &MockDB{}
	for i := 0; i < 20; i++ {
		mdb.push(fmt.Sprintf("pupa-test-%d", i), i)
	}
	return mdb
}

func (m *MockDB) push(text string, expiration int) *Note {
	created := time.Now()
	uid, _ := uuid.NewV4()
	note := &Note{
		ID:         uid,
		Text:       text,
		Created:    created,
		Expiration: expiration,
	}
	m.Notes = append(m.Notes, note)
	return note
}
