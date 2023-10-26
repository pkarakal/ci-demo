package router

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"gitlab.com/pkarakal/demo/pkg/models"
	"go.uber.org/zap"
	"gorm.io/gorm"
)

type UserCreationInput struct {
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
	Email      string `json:"email"`
}

func createUserRouter(g *gin.RouterGroup) {
	users := g.Group("/users")
	users.GET("", usersHandler)
	users.POST("", createUser)
	users.GET("/:id", getUserByID)
	users.GET("/:id/todos", getUserTodos)
}

func usersHandler(ctx *gin.Context) {
	db := ctx.MustGet("db").(*gorm.DB)
	var users []models.User

	if err := db.Model(&models.User{}).Preload("Todos").Find(&users).Error; err != nil {
		zap.L().Error("No user could be found", zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No user could be found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": users,
	})
}

func getUserByID(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		zap.L().Error("Failed to parse the id. ID must be an integer", zap.Error(err))
		ctx.JSON(http.StatusBadGateway, "User id must be an integer")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	var user models.User

	if err := db.Model(&models.User{}).Preload("Todos").First(&user, id).Error; err != nil {
		zap.L().Error("User with specified id could not be found", zap.Uint64("id", id), zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("User with id %d was not found", id)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": user,
	})
}

func getUserTodos(ctx *gin.Context) {
	userID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		zap.L().Error("Failed to parse the id. ID must be an integer", zap.Error(err))
		ctx.JSON(http.StatusBadGateway, "User id must be an integer")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)

	zap.L().Info(fmt.Sprintf("%v %v", userID, db))

	var todos []models.Todo

	res := db.Model(&models.Todo{}).Where("user_id = ?", userID).Take(&todos)

	if res.Error != nil {
		zap.L().Error("Couldn't get user todos", zap.Uint64("userID", userID), zap.Error(res.Error))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("couldn't get todo items for user %d", userID).Error()})
		return
	}

	ctx.JSON(http.StatusOK, todos)
}

func createUser(ctx *gin.Context) {
	var req UserCreationInput

	if err := ctx.BindJSON(&req); err != nil {
		zap.L().Error("Couldn't parse json input as UserCreationInput", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, fmt.Errorf("couldn't parse input %v", err))
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	var user models.User
	res := db.Model(&models.User{}).Where("email = ?", req.Email).First(&user)
	if res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		zap.L().Error("Couldn't fetch users from db", zap.Error(res.Error))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Errorf("couldn't fetch users from db %v", res.Error)})
		return
	}

	if res.RowsAffected > 0 {
		zap.L().Error("Duplicate email sent in request body")
		ctx.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("User with email %s already exists", req.Email)})
		return
	}

	user = models.User{
		GivenName:  req.GivenName,
		FamilyName: req.FamilyName,
		Email:      req.Email,
	}

	result := db.Model(&models.User{}).Create(&user)
	if result.Error != nil {
		zap.L().Error("Couldn't create user", zap.Error(result.Error))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't create user"})
		return
	}
	ctx.JSON(http.StatusOK, user)
}
