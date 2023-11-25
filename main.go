package main

import (
	"fmt"

	"gitlab.com/pkarakal/demo/internal/config"
	"gitlab.com/pkarakal/demo/pkg/models"
	"gitlab.com/pkarakal/demo/pkg/router"
	"go.uber.org/zap"
)

func main() {
	logger, undo := config.InitLogging()
	defer func(logger *zap.Logger) {
		err := logger.Sync()
		if err != nil {
			fmt.Printf("error syncing logger, %v\n", err)
		}
	}(logger)
	defer undo()
	c, err := config.LoadConfig()
	if err != nil {
		logger.Fatal("couldn't load configuration. Terminating", zap.Error(err))
	}
	db, err := c.OpenDatabase()
	if err != nil {
		logger.Fatal("couldn't connect to database. Terminating", zap.Error(err))
	}
	_ = models.Migrate(db)
	logger.Debug(fmt.Sprintf("Successfully connected to the database %v", db))
	e, v1 := router.InitRouter(db)
	router.Routes(v1)
	_ = e.Run(fmt.Sprintf(":%d", c.Settings.Port))
}
