package mongorepo

import (
	"encoding/json"
	"fmt"
	"strings"

	"go.mongodb.org/mongo-driver/bson"
)

type SimpleQuery struct {
	Query      string
	Projection string
	Sort       string
	Pageable   [2]int
}

func (q *SimpleQuery) ToClassicQuery(params ...interface{}) (ClassicQuery, error) {
	classic := ClassicQuery{
		Pageable: q.Pageable,
	}

	if q.Query != "" {
		queryStr := q.replaceParams(q.Query, params...)
		err := bson.UnmarshalExtJSON([]byte(queryStr), true, &classic.Query)
		if err != nil {
			return ClassicQuery{}, fmt.Errorf("error parsing query: %w", err)
		}
	}

	if q.Projection != "" {
		err := bson.UnmarshalExtJSON([]byte(q.Projection), true, &classic.Projection)
		if err != nil {
			return ClassicQuery{}, fmt.Errorf("error parsing projection: %w", err)
		}
	}

	if q.Sort != "" {
		var sortMap map[string]interface{}
		err := json.Unmarshal([]byte(q.Sort), &sortMap)
		if err != nil {
			return ClassicQuery{}, fmt.Errorf("error parsing sort: %w", err)
		}
		for k, v := range sortMap {
			classic.Sort = append(classic.Sort, bson.E{Key: k, Value: v})
		}
	}

	return classic, nil
}

func (q *SimpleQuery) replaceParams(query string, params ...interface{}) string {
	for i, param := range params {
		placeholder := fmt.Sprintf("?%d", i+1)
		replacement := fmt.Sprintf("%v", param)
		query = strings.Replace(query, placeholder, replacement, -1)
	}
	return query
}
