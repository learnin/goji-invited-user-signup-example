package main

import (
	"github.com/learnin/goji-invited-user-signup-example/helpers"
	"github.com/learnin/goji-invited-user-signup-example/models"
)

func main() {
	var ds helpers.DataSource
	if err := ds.Connect(); err != nil {
		panic(err)
	}
	defer ds.Close()
	db := ds.GetDB()
	db.LogMode(true)
	if d := db.AutoMigrate(&models.InviteUser{}); d.Error != nil {
		panic(d.Error)
	}
}
