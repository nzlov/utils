package db

import (
	"context"

	"github.com/vikstrous/dataloadgen"
)

func Loader[T any](column string, key func(T) string, options ...dataloadgen.Option) *dataloadgen.Loader[string, T] {
	return dataloadgen.NewLoader(func(ctx context.Context, s []string) ([]T, []error) {
		db := For(ctx)
		objs := []T{}
		rs := make([]T, len(s))
		es := make([]error, len(s))
		err := db.Where(column+" in (?)", s).Find(&objs).Error
		rm := map[string]T{}
		for _, v := range objs {
			rm[key(v)] = v
		}
		for i, v := range s {
			rs[i] = rm[v]
			es[i] = err
		}
		return rs, es
	})
}
