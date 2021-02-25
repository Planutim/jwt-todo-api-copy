package data

import (
	"context"
	"os"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Todo struct {
	UserID primitive.ObjectID `json:"user_id" bson:"user_id, omitempty"`
	Title  string             `json:"title"`
}

type Todos struct {
}

func NewTodos() *Todos {
	return &Todos{}
}

func (t *Todos) Save(td *Todo) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("mongodb")))
	if err != nil {
		return err
	}
	db := client.Database("jwt_users")
	todosCollection := db.Collection("todos")

	_, err = todosCollection.InsertOne(ctx, &td)

	if err != nil {
		return err
	}
	return nil
}

func (t *Todos) GetAll(userid primitive.ObjectID) ([]*Todo, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("mongodb")))
	if err != nil {
		return nil, err
	}
	db := client.Database("jwt_users")
	todosCollection := db.Collection("todos")

	cursor, err := todosCollection.Find(ctx, bson.M{
		"user_id": userid,
	})
	if err != nil {
		return nil, err
	}
	var todos []*Todo
	err = cursor.All(ctx, &todos)
	if err != nil {
		return nil, err
	}
	return todos, nil
}
