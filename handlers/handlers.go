package handlers

import (
	"fmt"
	"jwt_rewrite/data"
	"jwt_rewrite/helpers"
	"net/http"
	"os"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type JwtHandler struct {
	users  *data.Users
	todos  *data.Todos
	token  *helpers.Token
	client *helpers.RedisClient
}

func NewJwtHandler() (*JwtHandler, error) {
	client, err := helpers.NewRedisClient()
	if err != nil {
		return nil, err
	}
	users, err := data.NewUsers()
	if err != nil {
		return nil, err
	}
	return &JwtHandler{
		users:  users,
		todos:  data.NewTodos(),
		token:  helpers.NewToken(),
		client: client,
	}, nil
}

func (jh *JwtHandler) Register(c *gin.Context) {
	var u data.User

	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusUnprocessableEntity, fmt.Errorf("Register error\n%s", err.Error()))
		return
	}

	err := jh.users.RegisterUser(&u)
	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, fmt.Errorf("Error stuff\n %s", err.Error()))
	}
	c.JSON(http.StatusCreated, "created user successfully")
}
func (jh *JwtHandler) Login(c *gin.Context) {
	var u data.User

	if err := c.ShouldBindJSON(&u); err != nil {
		c.JSON(http.StatusUnprocessableEntity, fmt.Errorf("Herovii json\n %s", err.Error()))
		return
	}

	user, err := jh.users.GetUser(u.Username)
	if err != nil {

		c.JSON(http.StatusUnauthorized, fmt.Errorf("Nevernie detali logina\n %s", err.Error()))
		return
	}
	if user.Password != u.Password {
		c.JSON(http.StatusUnauthorized, "Nevernii parol")
		return
	}

	ts, err := jh.token.CreateToken(user.ID)

	if err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}

	saveErr := jh.client.CreateAuth(user.ID, ts)
	if saveErr != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
		return
	}
	tokens := map[string]string{
		"access_token":  ts.AccessToken,
		"refresh_token": ts.RefreshToken,
	}
	c.JSON(http.StatusOK, tokens)
}

func (jh *JwtHandler) Refresh(c *gin.Context) {
	mapToken := map[string]string{}
	if err := c.ShouldBindJSON(&mapToken); err != nil {
		c.JSON(http.StatusUnprocessableEntity, err.Error())
	}
	refreshToken := mapToken["refresh_token"]

	token, err := jwt.Parse(refreshToken, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return []byte(os.Getenv("REFRESH_SECRET")), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, "Refresh token expired")
		return
	}

	// is token valid?
	if _, ok := token.Claims.(jwt.Claims); !ok && !token.Valid {
		c.JSON(http.StatusUnauthorized, err)
	}
	//token is valid
	claims, ok := token.Claims.(jwt.MapClaims)
	if ok && token.Valid {
		refreshUuid, ok := claims["refresh_uuid"].(string)
		//convert the interface to string
		if !ok {
			c.JSON(http.StatusUnprocessableEntity, err)
			return
		}
		// userId, err := strconv.ParseUint(fmt.Sprintf("%.f", claims["user_id"]), 10, 64)
		userId := primitive.NewObjectID()
		if err != nil {
			c.JSON(http.StatusUnprocessableEntity, "Error occured")
			return
		}
		//Delete the previous Refresh token
		deleted, delErr := jh.client.DeleteAuth(refreshUuid)
		if delErr != nil || deleted == 0 {
			c.JSON(http.StatusUnauthorized, "unauthorized")
			return
		}
		//Create new pairs of refresh and access tokens
		ts, createErr := jh.token.CreateToken(userId)
		if createErr != nil {
			c.JSON(http.StatusForbidden, createErr.Error())
			return
		}
		//save the tokens metadata to redis
		saveErr := jh.client.CreateAuth(userId, ts)
		if saveErr != nil {
			c.JSON(http.StatusForbidden, saveErr.Error())
			return
		}

		tokens := map[string]string{
			"access_token":  ts.AccessToken,
			"refresh_token": ts.RefreshToken,
		}
		c.JSON(http.StatusCreated, tokens)
	} else {
		c.JSON(http.StatusUnauthorized, "refresh expired")
	}
}

func (jh *JwtHandler) TokenAuthMiddleWare() gin.HandlerFunc {
	return func(c *gin.Context) {
		err := jh.token.TokenValid(c.Request)
		if err != nil {
			c.JSON(http.StatusUnauthorized, err.Error())
			c.Abort()
			return
		}
		c.Next()
	}
}

func (jh *JwtHandler) CreateTodo(c *gin.Context) {
	var td *data.Todo
	if err := c.ShouldBindJSON(&td); err != nil {
		c.JSON(http.StatusUnprocessableEntity, "invalid json")
		return
	}
	tokenAuth, err := jh.token.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "ERR FROM EXTRACTTOKEN")
		return
	}
	userId, err := jh.client.FetchAuth(tokenAuth)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "ERR FROM FETCHAUTH")
		return
	}
	td.UserID = userId
	//save the todo to a database
	err = jh.todos.Save(td)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusCreated, td)

}

func (jh *JwtHandler) ListTodo(c *gin.Context) {
	tokenAuth, err := jh.token.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, err)
		return
	}
	_, err = jh.client.FetchAuth(tokenAuth)
	if err != nil {
		c.JSON(http.StatusUnauthorized, err)
		return
	}
	todos, err := jh.todos.GetAll(tokenAuth.UserId)
	if err != nil {
		c.JSON(http.StatusBadRequest, err)
		return
	}
	c.JSON(http.StatusOK, todos)
}

func (jh *JwtHandler) Logout(c *gin.Context) {
	au, err := jh.token.ExtractTokenMetadata(c.Request)
	if err != nil {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	deleted, delErr := jh.client.DeleteAuth(au.AccessUuid)
	if delErr != nil || deleted == 0 {
		c.JSON(http.StatusUnauthorized, "unauthorized")
		return
	}
	c.JSON(http.StatusOK, "Successfully logged out")
}
