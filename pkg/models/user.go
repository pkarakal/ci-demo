package models

import "gorm.io/gorm"

type User struct {
	gorm.Model
	GivenName  string `json:"givenName"`
	FamilyName string `json:"familyName"`
	Email      string `json:"email" gorm:"unique"`
	Todos      []Todo `json:"todos"`
}
