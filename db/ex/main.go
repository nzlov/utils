package main

import (
	"context"
	"fmt"
	"time"

	"github.com/nzlov/utils/db"
	"github.com/vikstrous/dataloadgen"
)

type A struct {
	db.Model

	AID  string
	Name string
	S    db.Array[string]       `gorm:"serializer:json"`
	M    db.Map[string, string] `gorm:"serializer:json"`
}

func AGet(ctx context.Context, id string) (A, error) {
	db := db.For(ctx)

	obj := A{}
	return obj, db.Where("name = ?", id).First(&obj).Error
}

type B struct {
	db.Model

	AID  string
	Name string
	Age  int64
}

func BAdd(ctx context.Context, req *B) error {
	return db.For(ctx).Create(req).Error
}

func BGet(ctx context.Context, id string) (B, error) {
	db := db.For(ctx)

	obj := B{}
	return obj, db.Where("name = ?", id).First(&obj).Error
}

func BF(b B) []string { return []string{b.AID, b.Name} }

var Loader = db.Loaders([]string{"a_id", "name"}, BF, dataloadgen.WithWait(time.Second))

func main() {
	cfg := db.Config{
		Driver: "sqlite",
		URL:    ":memory:",
	}

	if err := cfg.DB().AutoMigrate(new(A), new(B)); err != nil {
		panic(err)
	}

	ctx := cfg.Ctx(context.Background())

	if err := cfg.DB().Create(&A{
		AID:  "a",
		Name: "a",
		S:    db.Array[string]{"s1", "s2"},
		M:    db.Map[string, string]{"m1": "m1", "m2": "m2"},
	}).Error; err != nil {
		panic(err)
	}

	if a, err := AGet(ctx, "a"); err != nil {
		panic(err)
	} else {
		fmt.Println(a)
	}

	if err := cfg.DB().Create(&B{
		AID:  "a",
		Name: "a",
		Age:  1,
	}).Error; err != nil {
		panic(err)
	}
	if err := cfg.DB().Create(&B{
		AID:  "a",
		Name: "b",
		Age:  2,
	}).Error; err != nil {
		panic(err)
	}
	if err := cfg.DB().Create(&B{
		AID:  "b",
		Name: "b",
		Age:  3,
	}).Error; err != nil {
		panic(err)
	}
	if err := cfg.DB().Create(&B{
		AID:  "b",
		Name: "a",
		Age:  4,
	}).Error; err != nil {
		panic(err)
	}

	t := func(a ...string) {
		v, err := Loader.Load(ctx, db.LoadersKey(a...))
		if err != nil {
			fmt.Println(a, err)
		} else {
			fmt.Println(a, v.AID, v.Name, v.Age)
		}
	}
	go t("a", "b")
	go t("a", "1")
	go t("b", "b")
	go t("b", "a")
	go t("b", "a", "b")
	select {}
}
