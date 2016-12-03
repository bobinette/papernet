package gin

import (
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/dgrijalva/jwt-go"
	"github.com/gin-gonic/gin"

	"github.com/bobinette/papernet"
	"github.com/bobinette/papernet/oauth"
)

type AuthHandler struct {
	GoogleClient *oauth.GoogleOAuthClient
	Repository   papernet.UserRepository

	SigningKey papernet.SigningKey
}

func (h *AuthHandler) RegisterRoutes(router *gin.Engine) {
	router.GET("/api/auth", h.AuthURL)
	router.GET("/api/auth/google", h.Google)
	router.GET("/api/me", h.Me)
	router.POST("/api/bookmarks", h.Me)
}

func (h *AuthHandler) Me(c *gin.Context) {
	authHeader, ok := c.Request.Header["Authorization"]
	if !ok || len(authHeader) != 1 {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "No token found",
		})
		return
	}

	token := authHeader[0]
	if !strings.HasPrefix(token, "Bearer ") {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "No bearer",
		})
		return
	}

	token = token[len("Bearer "):]
	userID, err := decodeToken(h.SigningKey.Key, token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	user, err := h.Repository.Get(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"data": user,
	})
}

func (h *AuthHandler) AuthURL(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]interface{}{
		"url": h.GoogleClient.LoginURL(),
	})
}

func (h *AuthHandler) Google(c *gin.Context) {
	state := c.Query("state")
	code := c.Query("code")

	user, err := h.GoogleClient.ExchangeToken(state, code)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	err = h.Repository.Upsert(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	token, err := encodeToken(h.SigningKey.Key, user)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"access_token": token,
	})
}

func (h *AuthHandler) Bookmarks(c *gin.Context) {
	authHeader, ok := c.Request.Header["Authorization"]
	if !ok || len(authHeader) != 1 {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": "No token found",
		})
		return
	}

	token := authHeader[0]
	if !strings.HasPrefix(token, "Bearer ") {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": "No bearer",
		})
		return
	}

	token = token[len("Bearer "):]
	userID, err := decodeToken(h.SigningKey.Key, token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	user, err := h.Repository.Get(userID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	var payload struct {
		Add    []int `json:"add"`
		Remove []int `json:"remove"`
	}
	err = c.BindJSON(&payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	in := func(i int, a []int) bool {
		for _, e := range a {
			if e == i {
				return true
			}
		}
		return false
	}

	bookmarks := make(map[int]struct{})
	for _, b := range user.Bookmarks {
		if !in(b, payload.Remove) {
			bookmarks[b] = struct{}{}
		}
	}

	for _, b := range payload.Add {
		if !in(b, payload.Remove) {
			bookmarks[b] = struct{}{}
		}
	}

	user.Bookmarks = func(m map[int]struct{}) []int {
		i := 0
		a := make([]int, len(m))
		for k, _ := range m {
			a[i] = k
			i++
		}
		return a
	}(bookmarks)

	err = h.Repository.Upsert(user)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"user": user,
	})
}

type papernetClaims struct {
	UserID string `json:"user_id"`
	jwt.StandardClaims
}

func encodeToken(key string, user *papernet.User) (string, error) {
	// Create the Claims
	claims := papernetClaims{
		user.ID,
		jwt.StandardClaims{
			ExpiresAt: time.Now().AddDate(0, 2, 0).Unix(),
			Issuer:    "papernet",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(key))
}

func decodeToken(key, tokenString string) (string, error) {
	claims := papernetClaims{}

	token, err := jwt.ParseWithClaims(tokenString, &claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(key), nil
	})
	if err != nil {
		return "", err
	}

	if claims, ok := token.Claims.(*papernetClaims); ok && token.Valid {
		return claims.UserID, nil
	}

	return "", errors.New("could not get claims claims")
}
