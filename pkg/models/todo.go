package models

import "gorm.io/gorm"

type Todo struct {
	gorm.Model
	UserID      uint   `json:"-"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Completed   bool   `json:"completed"`
}
