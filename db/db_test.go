package db

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

type dbAllTestModel struct {
	Model
	Name string
}

// openTestDB 使用共享内存数据库隔离测试数据，避免文件状态影响分页与并发断言。
func openTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dbName := fmt.Sprintf("file:dball_%d?mode=memory&cache=shared", time.Now().UnixNano())
	db, err := gorm.Open(sqlite.Open(dbName), &gorm.Config{})
	if err != nil {
		t.Fatalf("打开测试数据库失败: %v", err)
	}
	if err := db.AutoMigrate(&dbAllTestModel{}); err != nil {
		t.Fatalf("迁移测试表失败: %v", err)
	}

	return db
}

// seedTestModels 统一构造连续主键数据，方便验证分页是否完整覆盖全部记录。
func seedTestModels(t *testing.T, db *gorm.DB, total int) {
	t.Helper()

	models := make([]dbAllTestModel, 0, total)
	for i := 0; i < total; i++ {
		models = append(models, dbAllTestModel{Name: fmt.Sprintf("item-%d", i+1)})
	}
	if err := db.Create(&models).Error; err != nil {
		t.Fatalf("写入测试数据失败: %v", err)
	}
}

// testCtx 将测试数据库注入上下文，确保 DBAll 与真实使用方式一致地通过 For(ctx) 取库。
func testCtx(db *gorm.DB) context.Context {
	return Ctx(context.Background(), db)
}

// TestDBAllPagination 避免固定页大小变更后漏数或重复消费，保证分页遍历完整性。
func TestDBAllPagination(t *testing.T) {
	db := openTestDB(t)
	seedTestModels(t, db, 250)

	seen := make(map[uint]int, 250)
	var mu sync.Mutex

	err := DBAll(testCtx(db), func(ctx context.Context, item dbAllTestModel) error {
		mu.Lock()
		defer mu.Unlock()
		seen[item.ID]++
		return nil
	}, WithDBAllPageSize(80), WithDBAllConcurrency(4))
	if err != nil {
		t.Fatalf("遍历数据失败: %v", err)
	}

	if len(seen) != 250 {
		t.Fatalf("处理记录数不正确: got=%d want=250", len(seen))
	}
	for id := uint(1); id <= 250; id++ {
		if seen[id] != 1 {
			t.Fatalf("记录处理次数异常: id=%d count=%d", id, seen[id])
		}
	}
}

// TestDBAllConcurrency 确认 worker 数量配置生效，避免回调仍被串行执行导致吞吐没有提升。
func TestDBAllConcurrency(t *testing.T) {
	db := openTestDB(t)
	seedTestModels(t, db, 8)

	var running int32
	var maxRunning int32

	err := DBAll(testCtx(db), func(ctx context.Context, item dbAllTestModel) error {
		current := atomic.AddInt32(&running, 1)
		defer atomic.AddInt32(&running, -1)

		for {
			max := atomic.LoadInt32(&maxRunning)
			if current <= max {
				break
			}
			if atomic.CompareAndSwapInt32(&maxRunning, max, current) {
				break
			}
		}

		// 短暂阻塞回调，让多个 worker 有机会重叠执行，便于稳定观察并发峰值。
		time.Sleep(40 * time.Millisecond)
		return nil
	}, WithDBAllPageSize(8), WithDBAllConcurrency(4))
	if err != nil {
		t.Fatalf("并发遍历失败: %v", err)
	}

	if atomic.LoadInt32(&maxRunning) < 2 {
		t.Fatalf("未观察到并发执行: max=%d", atomic.LoadInt32(&maxRunning))
	}
}

// TestDBAllCallbackError 保障回调报错后能向上返回，并尽快停止后续分页与消费。
func TestDBAllCallbackError(t *testing.T) {
	db := openTestDB(t)
	seedTestModels(t, db, 20)

	wantErr := errors.New("callback failed")
	var processed int32

	err := DBAll(testCtx(db), func(ctx context.Context, item dbAllTestModel) error {
		atomic.AddInt32(&processed, 1)
		if item.ID == 1 {
			return wantErr
		}
		return nil
	}, WithDBAllPageSize(1), WithDBAllConcurrency(1))
	if !errors.Is(err, wantErr) {
		t.Fatalf("返回错误不正确: got=%v want=%v", err, wantErr)
	}
	if atomic.LoadInt32(&processed) >= 20 {
		t.Fatalf("错误后未及时停止处理: processed=%d", atomic.LoadInt32(&processed))
	}
}

// TestDBAllContextCanceled 验证外部取消会透传到内部 goroutine，避免无意义继续查库和执行回调。
func TestDBAllContextCanceled(t *testing.T) {
	db := openTestDB(t)
	seedTestModels(t, db, 10)

	ctx, cancel := context.WithCancel(testCtx(db))
	cancel()

	var processed int32
	err := DBAll(ctx, func(ctx context.Context, item dbAllTestModel) error {
		atomic.AddInt32(&processed, 1)
		return nil
	}, WithDBAllPageSize(2), WithDBAllConcurrency(2))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("返回错误不正确: got=%v want=%v", err, context.Canceled)
	}
	if atomic.LoadInt32(&processed) != 0 {
		t.Fatalf("取消后不应继续处理数据: processed=%d", atomic.LoadInt32(&processed))
	}
}

// TestDBAllNilContext 避免调用方遗漏注入上下文时触发 panic，改为返回明确错误帮助快速定位问题。
func TestDBAllNilContext(t *testing.T) {
	err := DBAll[dbAllTestModel](nil, func(ctx context.Context, item dbAllTestModel) error {
		return nil
	})
	if !errors.Is(err, ErrNilContext) {
		t.Fatalf("返回错误不正确: got=%v want=%v", err, ErrNilContext)
	}
}
