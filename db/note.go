package db

import (
	"context"
	sq "github.com/Masterminds/squirrel"
	"github.com/gofrs/uuid"
	"github.com/jackc/pgx/v4"
	"time"
)

type Note struct {
	ID         uuid.UUID
	Text       string
	Created    time.Time
	Expiration int
}

type NoteHandler interface {
	Create(ctx context.Context, uid uuid.UUID, text string, exp_time int) (*Note, error)
	Get(ctx context.Context, uid uuid.UUID) (*Note, error)
	List(ctx context.Context) ([]*Note, error)
	Update(ctx context.Context, uid uuid.UUID, newText string) (*Note, error)
	Delete(ctx context.Context, uid uuid.UUID) (*Note, error)
	ClearExpired(ctx context.Context) error
}

type NoteDB struct {
	conn *pgx.Conn
}

func (ndb *NoteDB) Get(ctx context.Context, uid uuid.UUID) (*Note, error) {
	sql, args, err := sq.Select("id, text, created, expiration").From("notes").Where(sq.Eq{"id": uid}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	n := &Note{}
	if err = ndb.conn.QueryRow(ctx, sql, args...).Scan(&n.ID, &n.Text, &n.Created, &n.Expiration); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return n, nil
}

func (ndb *NoteDB) List(ctx context.Context) ([]*Note, error) {
	sql, args, err := sq.Select("id, text, created, expiration").From("notes").OrderBy("created").
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	var notes []*Note
	rows, err := ndb.conn.Query(ctx, sql, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		n := &Note{}
		if err = rows.Scan(&n.ID, &n.Text, &n.Created, &n.Expiration); err != nil {
			return nil, err
		}
		notes = append(notes, n)
	}

	return notes, nil
}

func (ndb *NoteDB) Update(ctx context.Context, uid uuid.UUID, newText string) (*Note, error) {
	sql, args, err := sq.Update("notes").Set("text", newText).Where(sq.Eq{"id": uid}).
		Suffix("RETURNING created").
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	n := &Note{
		ID:   uid,
		Text: newText,
	}
	if err = ndb.conn.QueryRow(ctx, sql, args...).Scan(&n.Created); err != nil {
		return nil, err
	}
	return n, err
}

func (ndb *NoteDB) Delete(ctx context.Context, uid uuid.UUID) (*Note, error) {
	sql, args, err := sq.Delete("notes").Where(sq.Eq{"id": uid}).Suffix("RETURNING id, text, created").
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	n := &Note{}
	if err = ndb.conn.QueryRow(ctx, sql, args...).Scan(&n.ID, &n.Text, &n.Created); err != nil {
		if err == pgx.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}

	return n, nil
}

func NewNoteDB(conn *pgx.Conn) NoteHandler {
	return &NoteDB{conn: conn}
}

func (ndb *NoteDB) Create(ctx context.Context, uid uuid.UUID, text string, exp_time int) (*Note, error) {
	now := time.Now()
	query, args, err := sq.Insert("notes").
		SetMap(map[string]interface{}{
			"id":         uid,
			"text":       text,
			"created":    now,
			"expiration": exp_time,
		}).PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return nil, err
	}

	_, err = ndb.conn.Exec(ctx, query, args...)
	if err != nil {
		return nil, err
	}

	return &Note{
		ID:         uid,
		Text:       text,
		Created:    now,
		Expiration: exp_time,
	}, nil
}

func (ndb *NoteDB) ClearExpired(ctx context.Context) error {
	sql, args, err := sq.Delete("notes").Where(sq.Lt{"created+(expiration*interval '1 minute')": time.Now()}).
		PlaceholderFormat(sq.Dollar).ToSql()
	if err != nil {
		return err
	}

	_, err = ndb.conn.Exec(ctx, sql, args...)
	if err != nil {
		return err
	}

	return nil
}

//type NoteMock struct {}
//
//func (n NoteMock) Create(uuid uuid.UUID, text string) (*Note, error) {
//	panic("implement me")
//}
//
//type Server struct {
//	NoteHandler NoteHandler
//}
//
//
//var s = Server{&NoteDB{conn: nil}}
//var mockS = Server{NoteHandler: NoteMock{}}
