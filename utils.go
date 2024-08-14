package mongorepo

import (
	"errors"
	"reflect"
	"strings"
)

func (r *MongoRepository[T]) setIdField() error {
	var dummy T
	t := reflect.TypeOf(dummy)
	if t.Kind() == reflect.Ptr {
		t = t.Elem()
	}

	for i := 0; i < t.NumField(); i++ {
		if tag := t.Field(i).Tag.Get("bson"); tag != "" {
			tags := strings.Split(tag, ",")
			for _, t := range tags {
				if strings.TrimSpace(t) == "_id" {
					r.idFieldIndex = i
					return nil
				}
			}
		}
	}

	return errors.New("type does not have a field with bson:\"_id\" tag")
}
