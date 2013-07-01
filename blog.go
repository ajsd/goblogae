package blog

import (
	"errors"
	"time"

	"appengine"
	"appengine/datastore"
)

const kind = "BlogEntry"

// BlogEntry
type BlogEntry struct {
	Markdown string         `datastore:"md,noindex" json:"md"`
	Updated  time.Time      `datastore:"dt" json:"updated"`
	ID       int64          `datastore:"-" json:"id"`
	key      *datastore.Key `datastore:"-" json:"-"`
}

// Creates a key which can be used for datastore ops.
// id is the StringID of a key (generated from, e.g., datastore.Put).
func createKey(c appengine.Context, id int64) *datastore.Key {
	return datastore.NewKey(c, kind, "", id, nil)
}

func createIncompleteKey(c appengine.Context) *datastore.Key {
	return datastore.NewIncompleteKey(c, kind, nil)
}

func NewEntry(markdown string) *BlogEntry {
	return &BlogEntry{
		Markdown: markdown,
	}
}

func GetEntry(c appengine.Context, id int64) (*BlogEntry, error) {
	e := new(BlogEntry)
	key := createKey(c, id)
	if err := datastore.Get(c, key, e); err != nil {
		return nil, err
	}
	e.setKey(key)
	return e, nil
}

func DeleteEntry(c appengine.Context, id int64) error {
	return datastore.Delete(c, createKey(c, id))
}

func DeleteEntries(c appengine.Context, ids []int64) error {
	var keys []*datastore.Key
	for _, id := range ids {
		keys = append(keys, createKey(c, id))
	}
	return datastore.DeleteMulti(c, keys)
}

func (e *BlogEntry) Key(c appengine.Context) *datastore.Key {
	if e.key == nil && e.ID > 0 {
		e.key = createKey(c, e.ID)
	}
	return e.key
}

func (e *BlogEntry) SaveNew(c appengine.Context) error {
	e.Updated = time.Now()
	if key, err := datastore.Put(c, createIncompleteKey(c), e); err != nil {
		return err
	} else {
		e.setKey(key)
	}
	return nil
}

func (e *BlogEntry) Update(c appengine.Context) error {
	e.Updated = time.Now()
	if !e.hasKey() {
		return errors.New("Cannot update an incomplete entity. Use `SaveNew` instead.")
	}
	_, err := datastore.Put(c, e.Key(c), e)
	return err
}

func (e *BlogEntry) Delete(c appengine.Context) error {
	return datastore.Delete(c, e.Key(c))
}

func (e *BlogEntry) setKey(key *datastore.Key) {
	e.key = key
	e.ID = e.key.IntID()
}

func (e *BlogEntry) hasKey() bool {
	return e.key != nil || e.ID > 0
}

// Queries
// TODO(arunjit): Query cursors instead of Range, q.Run() instead of q.GetAll()

func runQuery(c appengine.Context, q *datastore.Query, r *Range) ([]BlogEntry, error) {
	if r != nil && r.Limit > 0 {
		q = q.Offset(r.Offset).Limit(r.Limit)
	}
	var es []BlogEntry
	keys, err := q.GetAll(c, &es)

	// Can't use range; it creates a copy of the struct instead of a reference to it.
	for i := 0; i < len(es); i++ {
		es[i].setKey(keys[i])
	}
	return es, err
}

func GetEntries(c appengine.Context, r *Range) ([]BlogEntry, error) {
	return runQuery(c, datastore.NewQuery(kind).Order("-dt"), r)
}

func QueryEntries(c appengine.Context, from, to string, r *Range) ([]BlogEntry, error) {
	q := datastore.NewQuery(kind).Order("-dt")
	if from != "" {
		t, err := parseTime(from)
		if err != nil {
			return nil, err
		}
		q = q.Filter("dt >=", t)
	}
	if to != "" {
		t, err := parseTime(to)
		if err != nil {
			return nil, err
		}
		q = q.Filter("dt <=", t)
	}

	return runQuery(c, q, r)
}

func parseTime(t string) (time.Time, error) {
	return time.Parse("2006-01-02", t)
}
