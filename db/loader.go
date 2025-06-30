package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/vikstrous/dataloadgen"
)

func Loader[T any](column string, key func(T) string, options ...dataloadgen.Option) *dataloadgen.Loader[string, T] {
	return dataloadgen.NewLoader(func(ctx context.Context, keys []string) ([]T, []error) {
		db := For(ctx)
		objs := []T{}
		rs := make([]T, len(keys))
		es := make([]error, len(keys))
		err := db.Where(column+" in (?)", keys).Find(&objs).Error
		rm := map[string]T{}
		for _, v := range objs {
			rm[key(v)] = v
		}
		for i, v := range keys {
			rs[i] = rm[v]
			es[i] = err
		}
		return rs, es
	}, options...)
}

func LoadersKey(keys ...string) string {
	return strings.Join(keys, "@nzlov@")
}

var (
	ErrNotFound = fmt.Errorf("not found")
	ErrKey      = fmt.Errorf("invalid key format")
)

func Loaders[T any](columns []string, key func(T) []string, options ...dataloadgen.Option) *dataloadgen.Loader[string, T] {
	return dataloadgen.NewLoader(func(ctx context.Context, keys []string) ([]T, []error) {
		db := For(ctx)
		objs := []T{}
		rs := make([]T, len(keys))
		es := make([]error, len(keys))

		css := make([][]string, len(columns))

		fmt.Println(keys)
		for i, c := range keys {
			cs := strings.Split(c, "@nzlov@")
			fmt.Println(cs)
			if len(cs) != len(columns) {
				es[i] = fmt.Errorf("%w: %s", ErrKey, cs)
				continue
			}
			for j, c := range cs {
				css[j] = append(css[j], c)
			}
		}
		for i, v := range columns {
			db = db.Where(v+" in (?)", css[i])
		}
		err := db.Find(&objs).Error
		rm := map[string]T{}
		for _, v := range objs {
			rm[strings.Join(key(v), "@nzlov@")] = v
		}
		if err != nil {
			for i := range keys {
				if es[i] != nil {
					continue
				}
				es[i] = err
			}
		} else {
			var ok bool
			for i, v := range keys {
				if es[i] != nil {
					continue
				}
				rs[i], ok = rm[v]
				if !ok {
					es[i] = fmt.Errorf("%w: %s", ErrNotFound, strings.Split(v, "@nzlov@"))
				}
			}
		}
		return rs, es
	}, options...)
}
