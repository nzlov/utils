package db

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/glebarez/sqlite"
	"golang.org/x/sync/errgroup"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var ErrNilContext = errors.New("nil context")

type Model struct {
	CreatedAt time.Time
	UpdatedAt time.Time
	DeletedAt gorm.DeletedAt `gorm:"index"`

	ID uint `gorm:"primarykey"`
}

func (d Model) GetID() uint {
	return d.ID
}

type SQLExecutionHistory struct {
	Model
	FileName string `gorm:"uniqueIndex"`
}

type Config struct {
	db *gorm.DB

	Driver string `json:"driver" yaml:"driver"`
	URL    string `json:"url"    yaml:"url"`
}

func (c *Config) DB() *gorm.DB {
	if c.db != nil {
		return c.db
	}
	db, err := c.Open()
	if err != nil {
		panic(err)
	}
	return db
}

type ctxKey struct{}

var _ctxKey = ctxKey{}

func For(ctx context.Context) *gorm.DB {
	return ctx.Value(_ctxKey).(*gorm.DB).Session(&gorm.Session{})
}

func (c *Config) Ctx(ctx context.Context) context.Context {
	return context.WithValue(ctx, _ctxKey, c.DB())
}

func Ctx(ctx context.Context, db *gorm.DB) context.Context {
	return context.WithValue(ctx, _ctxKey, db)
}

func CtxNew(ctx context.Context) context.Context {
	return context.WithValue(ctx, _ctxKey, For(ctx).Session(&gorm.Session{
		NewDB:       true,
		Initialized: true,
	}))
}

func Tx(ctx context.Context, f func(context.Context) error, opts ...*sql.TxOptions) error {
	return For(ctx).Transaction(func(tx *gorm.DB) error {
		return f(Ctx(ctx, tx))
	}, opts...)
}

type Option func(*gorm.Config)

func NewLoggerOp(logger logger.Interface) func(*gorm.Config) {
	return func(cfg *gorm.Config) {
		cfg.Logger = logger
	}
}

func (d *Config) Open(ops ...Option) (*gorm.DB, error) {
	cfg := gorm.Config{
		TranslateError: true,
		Logger:         logger.Default.LogMode(logger.Info),
	}
	for _, op := range ops {
		op(&cfg)
	}
	var err error
	switch d.Driver {
	case "postgres":
		d.db, err = gorm.Open(postgres.Open(d.URL), &cfg)
	case "sqlite":
		d.db, err = gorm.Open(sqlite.Open(d.URL), &cfg)
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

// ExecuteSQLFilesFromEmbed reads SQL files from an embedded directory and executes them if not already executed.
// Uses a transaction to ensure atomicity of SQL execution and history recording.
func ExecuteSQLFilesFromEmbed(ctx context.Context, fs embed.FS, dir string) error {
	db := For(ctx)

	// Auto-migrate the SQL execution history table
	if err := db.AutoMigrate(&SQLExecutionHistory{}); err != nil {
		return fmt.Errorf("failed to migrate SQL execution history table: %v", err)
	}

	// Read all SQL files from the embedded directory
	files, err := fs.ReadDir(dir)
	if err != nil {
		return fmt.Errorf("failed to read embedded directory: %v", err)
	}

	for _, file := range files {
		if filepath.Ext(file.Name()) != ".sql" {
			continue
		}

		// Check if the file has already been executed
		var history SQLExecutionHistory
		if err := db.Where("file_name = ?", file.Name()).First(&history).Error; err == nil {
			continue // Skip if already executed
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("failed to query execution history: %v", err)
		}

		// Read the SQL file content
		content, err := fs.ReadFile(dir + "/" + file.Name())
		if err != nil {
			return fmt.Errorf("failed to read SQL file: %v", err)
		}

		// Execute the SQL and record history in a transaction
		err = db.Transaction(func(tx *gorm.DB) error {
			sqlLines := strings.Split(string(content), "\n")
			for n, line := range sqlLines {
				if strings.TrimSpace(line) == "" {
					continue // Skip empty lines
				}
				if err := tx.Exec(line).Error; err != nil {
					return fmt.Errorf("failed to execute SQL line in file %s:%d, %v", file.Name(), n, err)
				}
			}

			if err := tx.Create(&SQLExecutionHistory{
				FileName: file.Name(),
			}).Error; err != nil {
				return fmt.Errorf("failed to record SQL execution: %v", err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}

	return nil
}

type DBModelID interface {
	GetID() uint
}

type DBAllConfig struct {
	PageSize    int
	Concurrency int
}

type DBAllOption func(*DBAllConfig)

// WithDBAllPageSize 允许按需调整分页大小，避免不同表规模下固定分页带来额外压力。
func WithDBAllPageSize(size int) DBAllOption {
	return func(cfg *DBAllConfig) {
		if size > 0 {
			cfg.PageSize = size
		}
	}
}

// WithDBAllConcurrency 允许按需调整回调并发数，兼顾吞吐和下游处理能力。
// channel 缓冲区会与并发数保持一致，避免生产速度远高于消费速度时积压过多内存。
func WithDBAllConcurrency(concurrency int) DBAllOption {
	return func(cfg *DBAllConfig) {
		if concurrency > 0 {
			cfg.Concurrency = concurrency
		}
	}
}

// DBAll 按主键递增分页读取数据，并将结果分发给固定数量的工作协程并发执行回调。
// 通过透传 context 可以让外部取消信号和内部错误更快地停止后续分页与回调处理。
func DBAll[T DBModelID](ctx context.Context, cb func(context.Context, T) error, ops ...DBAllOption) error {
	defaultConcurrency := runtime.NumCPU() * 2
	if defaultConcurrency <= 0 {
		defaultConcurrency = 1
	}

	cfg := DBAllConfig{
		PageSize:    100,
		Concurrency: defaultConcurrency,
	}
	for _, op := range ops {
		op(&cfg)
	}

	if ctx == nil {
		return ErrNilContext
	}

	db := For(ctx)
	g, ctx := errgroup.WithContext(ctx)
	lastid := uint(0)
	ch := make(chan T, cfg.Concurrency)

	for i := 0; i < cfg.Concurrency; i++ {
		g.Go(func() error {
			for v := range ch {
				if err := cb(ctx, v); err != nil {
					return err
				}
			}
			return nil
		})
	}

	g.Go(func() error {
		defer close(ch)

		for {
			if err := ctx.Err(); err != nil {
				return err
			}

			objs := make([]T, 0, cfg.PageSize)
			if err := db.Where("id > ?", lastid).Order("id").Limit(cfg.PageSize).Find(&objs).Error; err != nil {
				return err
			}
			if len(objs) == 0 {
				return nil
			}

			for _, v := range objs {
				lastid = v.GetID()
				select {
				case <-ctx.Done():
					return ctx.Err()
				case ch <- v:
				}
			}

			if len(objs) != cfg.PageSize {
				return nil
			}
		}
	})

	if err := g.Wait(); err != nil {
		return err
	}

	return nil
}
