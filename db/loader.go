package db

import (
	"context"
	"fmt"
	"strings"

	"github.com/vikstrous/dataloadgen"
	"gorm.io/gorm"
)

type whereKey string

var _whereKey whereKey = "nzlov@whereKey"

type where struct {
	Query string
	Args  []any
}

func WhereCtx(ctx context.Context, query string, args ...any) context.Context {
	return context.WithValue(ctx, _whereKey, &where{Query: query, Args: args})
}

func _Where(ctx context.Context) *where {
	w, ok := ctx.Value(_whereKey).(*where)
	if !ok {
		return nil
	}
	return w
}

func Loader[T any](column string, key func(T) string, options ...dataloadgen.Option) *dataloadgen.Loader[string, T] {
	return LoaderCtx(column, func(ctx context.Context, t T) string {
		return key(t)
	}, options...)
}

func LoaderCtx[T any](column string, key func(context.Context, T) string, options ...dataloadgen.Option) *dataloadgen.Loader[string, T] {
	return dataloadgen.NewLoader(func(ctx context.Context, keys []string) ([]T, []error) {
		db := gorm.G[T](For(ctx)).Scopes()
		rs := make([]T, len(keys))
		es := make([]error, len(keys))
		where := _Where(ctx)
		if where != nil {
			db = db.Where(where.Query, where.Args...)
		}
		objs, err := db.Where(column+" in (?)", keys).Find(ctx)
		if err != nil {
			for i := range keys {
				es[i] = err
			}
		} else {
			rm := map[string]T{}
			for _, v := range objs {
				rm[key(ctx, v)] = v
			}
			var ok bool
			for i, v := range keys {
				rs[i], ok = rm[v]
				if !ok {
					es[i] = fmt.Errorf("%w: %s", ErrNotFound, strings.Split(v, "@nzlov@"))
				}
			}
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
	return LoaderCtxs(columns, func(ctx context.Context, t T) []string {
		return key(t)
	}, options...)
}

func LoaderCtxs[T any](columns []string, key func(context.Context, T) []string, options ...dataloadgen.Option) *dataloadgen.Loader[string, T] {
	return dataloadgen.NewLoader(func(ctx context.Context, keys []string) ([]T, []error) {
		db := gorm.G[T](For(ctx)).Scopes()

		rs := make([]T, len(keys))
		es := make([]error, len(keys))

		css := make([][]string, len(columns))

		for i, c := range keys {
			cs := strings.Split(c, "@nzlov@")
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

		where := _Where(ctx)
		if where != nil {
			db = db.Where(where.Query, where.Args...)
		}

		objs, err := db.Find(ctx)

		if err != nil {
			for i := range keys {
				if es[i] != nil {
					continue
				}
				es[i] = err
			}
		} else {
			rm := map[string]T{}
			for _, v := range objs {
				rm[LoadersKey(key(ctx, v)...)] = v
			}
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
