package blog

import (
	"errors"
	"net/http"

	"github.com/ajsd/go/auth"
	"github.com/mjibson/appstats"
)

type BlogService struct {
	auth    auth.Authenticator
	authAll bool
}

func NewService() *BlogService {
	return &BlogService{}
}

func NewAuthenticatedService(auth auth.Authenticator, authAll bool) *BlogService {
	return &BlogService{auth, authAll}
}

// Search

type SearchArgs struct {
	Query string `json:"q"`
	Range *Range `json:"range,omitempty"`
}

type SearchResult struct {
	Entries []BlogEntry `json:"entries"`
}

func (s *BlogService) Search(r *http.Request, args *SearchArgs, result *SearchResult) error {
	if s.auth != nil && s.authAll && !s.auth.CheckAuth(r) {
		return auth.ErrForbidden
	}

	if args.Query == "" {
		return errors.New("Missing query")
	}

	c := appstats.NewContext(r)
	defer c.Save()

	p, err := parseQuery(args.Query)
	if err != nil {
		return err
	}

	// GetAll
	if p.All {
		if es, err := GetEntries(c, args.Range); err != nil {
			return err
		} else {
			result.Entries = es
		}
		return nil
	}

	// GetById
	if p.ID != 0 {
		if e, err := GetEntry(c, p.ID); err != nil {
			return err
		} else {
			result.Entries = []BlogEntry{*e}
		}
		return nil
	}

	// Query
	if es, err := QueryEntries(c, p.From, p.To, args.Range); err != nil {
		return err
	} else {
		result.Entries = es
	}
	return nil
}

// Save

type SaveArgs struct {
	ID       int64  `json:"id,omitempty"` // Required only for updating an existing entity
	Markdown string `json:"md"`
}

type SaveResult struct {
	Entry *BlogEntry `json:"entry"`
}

func (s *BlogService) Save(r *http.Request, args *SaveArgs, result *SaveResult) error {
	if s.auth != nil && !s.auth.CheckAuth(r) {
		return auth.ErrForbidden
	}

	if len(args.Markdown) == 0 {
		return errors.New("Missing markdown data")
	}
	// TODO(arunjit): Validate markdown?

	c := appstats.NewContext(r)
	defer c.Save()

	if args.ID != 0 {
		// Existing
		if e, err := GetEntry(c, args.ID); err != nil {
			return err
		} else {
			e.Markdown = args.Markdown
			if err := e.Update(c); err != nil {
				return err
			}
			result.Entry = e
		}
	} else {
		// New
		e := NewEntry(args.Markdown)
		if err := e.SaveNew(c); err != nil {
			return err
		}
		result.Entry = e
	}
	return nil
}

// Delete

type DeleteArgs struct {
	ID int64 `json:"id"`
}

type DeleteResult struct {
	ID int64 `json:"id"`
}

func (s *BlogService) Delete(r *http.Request, args *DeleteArgs, result *DeleteResult) error {
	if s.auth != nil && !s.auth.CheckAuth(r) {
		return auth.ErrForbidden
	}

	if args.ID == 0 {
		return errors.New("Missing ID")
	}

	c := appstats.NewContext(r)
	defer c.Save()

	if err := DeleteEntry(c, args.ID); err != nil {
		return err
	}
	result.ID = args.ID
	return nil
}
