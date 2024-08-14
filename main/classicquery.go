package mongorepo

import "go.mongodb.org/mongo-driver/bson"

type ClassicQuery struct {
	Query      bson.M
	Projection bson.M
	Sort       bson.D
	Pageable   [2]int
}
