package blog

import (
	"errors"
	"regexp"
	"strconv"
)

// Query parser

// Parse result
// All, ID are mutually exclusive to the rest and each other.
type parseResult struct {
	All      bool
	ID       int64
	From, To string
}

// Parse the query string
// The query will be like "*" => GetAll, "id:xxxxx" => GetById
// "from:yyyy-mm-dd to:yyyy-mm-dd" filtered by BlogEntry.Updated.
func parseQuery(q string) (*parseResult, error) {
	if q == "*" {
		return &parseResult{All: true}, nil
	}
	re := regexp.MustCompile(`(id|from|to):([a-zA-Z0-9_-]+)`)
	matches := re.FindAllStringSubmatch(q, -1)

	p := new(parseResult)
	for _, match := range matches {
		k, v := match[1], match[2]
		switch k {
		case "id":
			if id, err := strconv.ParseInt(v, 0, 64); err != nil {
				return nil, errors.New("Error parsing ID: " + err.Error())
			} else {
				return &parseResult{ID: id}, nil
			}
		case "from":
			p.From = v
		case "to":
			p.To = v
		}
	}
	return p, nil
}
