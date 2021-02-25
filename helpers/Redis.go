package helpers

import (
	"jwt_rewrite/data"
	"time"

	"github.com/go-redis/redis"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type RedisClient struct {
	client *redis.Client
}

func NewRedisClient() (*RedisClient, error) {
	dsn := "localhost:6379"
	client := redis.NewClient(&redis.Options{
		Addr: dsn,
	})
	_, err := client.Ping().Result()
	if err != nil {
		return nil, err
	}
	return &RedisClient{
		client: client,
	}, nil
}

func (rc *RedisClient) CreateAuth(userid primitive.ObjectID, td *TokenDetails) error {
	at := time.Unix(td.AtExpires, 0) // unix time back to utc
	rt := time.Unix(td.RtExpires, 0)

	now := time.Now()
	errAccess := rc.client.Set(td.AccessUuid, userid.Hex(), at.Sub(now)).Err()
	if errAccess != nil {
		return errAccess
	}
	errRefresh := rc.client.Set(td.RefreshUuid, userid.Hex(), rt.Sub(now)).Err()
	if errRefresh != nil {
		return errRefresh
	}
	return nil
}

func (rc *RedisClient) DeleteAuth(givenUuid string) (int64, error) {
	deleted, err := rc.client.Del(givenUuid).Result()
	if err != nil {
		return 0, err
	}
	return deleted, nil
}

func (rc *RedisClient) FetchAuth(authD *data.AccessDetails) (primitive.ObjectID, error) {
	userId, err := rc.client.Get(authD.AccessUuid).Result()
	if err != nil {
		return primitive.NewObjectID(), err
	}
	// userID, _ := strconv.ParseUint(userId, 10, 64)
	userID, err := primitive.ObjectIDFromHex(userId)
	if err != nil {
		return primitive.NewObjectID(), err
	}
	return userID, nil
}
