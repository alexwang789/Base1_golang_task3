package main

import (
	"gorm.io/gorm"
)

type students struct {
	ID    uint   // Standard field for the primary key
	Name  string // A regular string field
	Age   uint8  // An unsigned 8-bit integer
	Grade string
}

func Run(db *gorm.DB) {
	// Create a new user
	db.AutoMigrate(&students{})

	// Save the user to the database
	student := students{Name: "张三", Age: 20, Grade: "三年级"}
	db.Create(&student)
	db.Where("age >?", 18)
	db.Model(&student).Where("Name =?", "张三").Updates(students{Grade: "四年级"})
	db.Where("age <?", 15).Delete(&students{})

}
