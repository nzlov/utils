package db

import (
	"fmt"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DBModel struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	ID uint `gorm:"primarykey"`
}

func (d DBModel) GetID() uint {
	return d.ID
}

type DBConfig struct {
	db *gorm.DB

	Driver string `json:"driver" yaml:"driver"`
	URL    string `json:"url"    yaml:"url"`
}

func (d DBConfig) DB() *gorm.DB {
	if d.db != nil {
		return d.db
	}
	db, err := d.Open()
	if err != nil {
		panic(err)
	}
	return db
}

func (d DBConfig) Open() (*gorm.DB, error) {
	cfg := gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Info),
	}
	var err error
	switch d.Driver {
	case "postgres":
		d.db, err = gorm.Open(postgres.Open(d.URL), &cfg)
	case "mysql":
		d.db, err = gorm.Open(mysql.New(mysql.Config{
			DSN:               d.URL,
			DefaultStringSize: 256,
		}), &cfg)
	default:
		err = fmt.Errorf("not supported %s db driver", d.Driver)
	}
	return d.db, err
}

type DBModelID interface {
	GetID() uint
}

func DBAll[T DBModelID](db *gorm.DB, cb func(T) error) error {
	lastid := uint(0)

	objs := []T{}

	for {
		if err := db.Where("id > ?", lastid).Order("id").Limit(100).Find(&objs).Error; err != nil {
			return err
		}
		if len(objs) == 0 {
			break
		}
		for _, v := range objs {
			if err := cb(v); err != nil {
				return err
			}
			lastid = v.GetID()
		}
		if len(objs) != 100 {
			break
		}
	}
	return nil
}
