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

//User is a model
type User struct {
	ID       primitive.ObjectID `json:"id" bson:"_id, omitempty"`
	Username string             `json:"username" bson:"username", omitempty"`
	Password string             `json:"password" bson:"password, omitempty"`
}

//Users is wrapper for array of users
type Users struct {
}

//NewUsers creates new instance of Users
func NewUsers() (*Users, error) {
	return &Users{}, nil
}

//RegisterUser registers user
func (u *Users) RegisterUser(user *User) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("mongodb")))
	if err != nil {
		return err
	}
	defer client.Disconnect(ctx)

	user.ID = primitive.NewObjectID()
	db := client.Database("jwt_users")
	userCollection := db.Collection("users")

	_, err = userCollection.InsertOne(ctx, user)
	if err != nil {
		return err
	}
	return nil
}

//GetUser gets the user with given username
func (u *Users) GetUser(username string) (*User, error) {
	// for _, v := range users {
	// 	if v.Username == username {
	// 		return *v, nil
	// 	}
	// }
	// return User{}, errors.New("No user found")
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(os.Getenv("mongodb")))
	if err != nil {
		return nil, err
	}
	defer client.Disconnect(ctx)

	var user User
	db := client.Database("jwt_users")
	userCollection := db.Collection("users")

	cursor := userCollection.FindOne(ctx, bson.M{
		"username": username,
	})

	err = cursor.Decode(&user)
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// var users = []*User{
// 	{
// 		ID:       1,
// 		Username: "username",
// 		Password: "password",
// 	},
// 	{
// 		ID:       2,
// 		Username: "user1",
// 		Password: "parol123",
// 	},
// 	{
// 		ID:       3,
// 		Username: "Aiperi",
// 		Password: "Salieva95",
// 	},
// }
