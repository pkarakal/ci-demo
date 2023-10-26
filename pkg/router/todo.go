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

func createTodoRouter(g *gin.RouterGroup) {
	todos := g.Group("/todo")
	todos.GET("", todosHandler)
	todos.POST("", createTodo)
	todos.GET("/:id", todoHandler)
	todos.PATCH("/:id/markAsCompleted", markTodoAsCompleted)
}

type TodoCreationInput struct {
	UserID      uint   `json:"userID"`
	Title       string `json:"title"`
	Description string `json:"description"`
}

func todosHandler(ctx *gin.Context) {
	var todos []models.Todo

	db := ctx.MustGet("db").(*gorm.DB)
	if err := db.Model(&models.Todo{}).Find(&todos).Error; err != nil {
		zap.L().Error("No todo item could be found", zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": "No todo item could be found"})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": todos,
	})
}

func todoHandler(ctx *gin.Context) {
	id, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		zap.L().Error("Failed to parse the id. ID must be an integer", zap.Error(err))
		ctx.JSON(http.StatusBadGateway, "Todo id must be an integer")
		return
	}
	db := ctx.MustGet("db").(*gorm.DB)
	var todo models.Todo

	if err := db.Model(&models.Todo{}).First(&todo, id).Error; err != nil {
		zap.L().Error("Todo item with specified id could not be found", zap.Uint64("id", id), zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Todo item with id %d was not found", id)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": todo,
	})
}

func createTodo(ctx *gin.Context) {
	var req TodoCreationInput

	if err := ctx.BindJSON(&req); err != nil {
		zap.L().Error("Couldn't parse json input as TodoCreationInput", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("couldn't parse input %v", err).Error()})
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)

	var user models.User
	if err := db.Model(&models.User{}).Find(&user, req.UserID).Error; err != nil && errors.Is(err, gorm.ErrRecordNotFound) {
		zap.L().Error("Can't create a todo item for a non-existing user", zap.Error(err))
		ctx.JSON(http.StatusBadRequest, gin.H{"error": fmt.Errorf("can't create a todo for a user that doesn't exist")})
		return
	}

	todo := models.Todo{
		Title:       req.Title,
		Description: req.Description,
		Completed:   false,
		UserID:      req.UserID,
	}

	result := db.Model(&models.Todo{}).Create(&todo)
	if result.Error != nil {
		zap.L().Error("Couldn't create todo item", zap.Error(result.Error))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": "couldn't create todo item"})
		return
	}
	ctx.JSON(http.StatusOK, todo)
}

func markTodoAsCompleted(ctx *gin.Context) {
	todoID, err := strconv.ParseUint(ctx.Param("id"), 10, 64)
	if err != nil {
		zap.L().Error("Failed to parse the id. ID must be an integer", zap.Error(err))
		ctx.JSON(http.StatusBadGateway, "Todo id must be an integer")
		return
	}

	db := ctx.MustGet("db").(*gorm.DB)
	var todo models.Todo

	if err := db.Model(&models.Todo{}).First(&todo, todoID).Error; err != nil {
		zap.L().Error("Todo item with specified id could not be found", zap.Uint64("id", todoID), zap.Error(err))
		ctx.JSON(http.StatusNotFound, gin.H{"error": fmt.Sprintf("Todo item with id %d was not found", todoID)})
		return
	}

	todo.Completed = true

	if err := db.Model(&todo).Update("completed", true).Error; err != nil {
		zap.L().Error("Failed to mark todo item as completed", zap.Error(err))
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": fmt.Sprintf("Todo item with id %d couldn't be marked as completed", todoID)})
		return
	}

	ctx.JSON(http.StatusOK, gin.H{
		"data": todo,
	})
}
