package data

import "go.mongodb.org/mongo-driver/bson/primitive"

type AccessDetails struct {
	AccessUuid string
	UserId     primitive.ObjectID
}
