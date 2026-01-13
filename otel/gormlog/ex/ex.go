package main

import (
	"context"
	"fmt"
	"strings"

	"github.com/nzlov/utils/otel"
	"github.com/nzlov/utils/otel/gormlog"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type App struct{}

func (a *App) Run(ctx context.Context) error {
	db, err := gorm.Open(sqlite.Open("test.db"), &gorm.Config{
		Logger: gormlog.Default().LogMode(logger.Info),
	})
	if err != nil {
		panic("failed to connect database")
	}
	if err := db.AutoMigrate(new(A), new(B)); err != nil {
		panic(err)
	}
	{
		fmt.Println(db.Create(&A{
			Name: "a",
		}).Error)
		fmt.Println(db.Create(&B{
			Name: "b",
		}).Error)
	}

	{
		a := A{}
		{
			db.Where("name = ?", "a").First(&a)
			fmt.Println(a)
		}
		{
			db.Where("name = ?", "a").Find(&a)
			fmt.Println(a)
		}
		a.P = "100"
		db.Save(&a)
		{
			db.Where("name = ?", "a").First(&a)
			fmt.Println(a)
		}
		{
			db.Where("name = ?", "a").Find(&a)
			fmt.Println(a)
		}
	}
	{

		b := B{}
		db.Where("p = ?", "").Where("name = ?", "b").Find(&b)
		fmt.Println(b)
	}
	return nil
}

func (a *App) Shutdown(ctx context.Context) error {
	return nil
}

func main() {
	cfg := otel.Config{
		Name:          "tot",
		Type:          "httpc",
		LogSource:     true,
		MetricDisable: false,
		TraceDisable:  false,
	}
	if err := cfg.Run(new(App)); err != nil {
		panic(err)
	}
}

type A struct {
	gorm.Model

	Name string
	P    string
}

func (a *A) BeforeCreate(tx *gorm.DB) (err error) {
	a.P = "aaaa" + a.P
	return err
}

func (a *A) BeforeUpdate(tx *gorm.DB) (err error) {
	if !strings.HasPrefix(a.P, "aaaa") {
		a.P = "aaaa" + a.P
	}

	return err
}

func (a *A) AfterFind(tx *gorm.DB) (err error) {
	a.P = a.P[4:]
	return err
}

type B struct {
	gorm.Model

	Name string
	P    string
}
