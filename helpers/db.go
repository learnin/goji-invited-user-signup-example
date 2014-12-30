package helpers

import (
	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

type DataSource struct {
	db *gorm.DB
	tx *gorm.DB
}

func (ds *DataSource) Connect() error {
	db, err := gorm.Open("mysql", "example_app:example_app@tcp(localhost:3306)/invited_user_signup_example?charset=utf8&parseTime=True&loc=Asia%2FTokyo")
	if err != nil {
		return err
	}
	db.DB()
	db.LogMode(true)
	ds.db = &db
	return nil
}

func (ds *DataSource) Close() error {
	return ds.db.Close()
}

func (ds *DataSource) GetDB() *gorm.DB {
	return ds.db
}

func (ds *DataSource) GetTx() *gorm.DB {
	return ds.tx
}

func (ds *DataSource) DoInTransaction(callback func(ds *DataSource) error) error {
	ds.tx = ds.db.Begin()
	if err := callback(ds); err != nil {
		ds.tx.Rollback()
		return err
	}
	if err := ds.tx.Commit().Error; err != nil {
		return err
	}
	return nil
}
