package service

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
)

// TestScenicExtraCacheTTL 验证周边推荐 Redis 缓存的读写往返、TTL 与"空结果不缓存"三条约定。
// 依赖本机 Redis(REDIS_ADDR,默认 127.0.0.1:6379),不可用时跳过。
func TestScenicExtraCacheTTL(t *testing.T) {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		addr = "127.0.0.1:6379"
	}
	rdb := redis.NewClient(&redis.Options{Addr: addr})
	defer rdb.Close()

	ctx := context.Background()
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Skip("Redis 不可用")
	}

	const cacheTTL = 90 * time.Second
	svc := &ScenicExtraService{rdb: rdb, cacheTTL: cacheTTL}

	// 独立测试键前缀,避免污染真实缓存键(真实键为 scenic:around:{id}:{limit})
	key := "scenic:around:test-9999:0"
	emptyKey := "scenic:around:test-9999:1"
	defer rdb.Del(ctx, key, emptyKey)

	want := []AroundItem{
		{ID: "B001", Name: "人民公园", Type: "风景名胜;公园广场", Address: "成都市青羊区少城路12号", Distance: 120, Lat: 30.6571, Lng: 104.0552},
		{ID: "B002", Name: "宽窄巷子", Type: "风景名胜;风景名胜;景点", Address: "成都市青羊区金河路口", Distance: 460, Lat: 30.6695, Lng: 104.0576},
	}
	svc.setCache(ctx, key, want)

	got := svc.getCache(ctx, key)
	if len(got) != len(want) {
		t.Fatalf("缓存往返条目数=%d, want %d", len(got), len(want))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("第 %d 条缓存内容不一致: got %+v, want %+v", i, got[i], want[i])
		}
	}

	ttl, err := rdb.TTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("读取缓存 TTL 失败: %v", err)
	}
	if ttl <= 0 || ttl > cacheTTL {
		t.Errorf("TTL=%v, 期望落在 (0, %v]", ttl, cacheTTL)
	}

	// 空结果不写缓存
	svc.setCache(ctx, emptyKey, []AroundItem{})
	if n, err := rdb.Exists(ctx, emptyKey).Result(); err != nil {
		t.Fatalf("检查缓存键存在性失败: %v", err)
	} else if n != 0 {
		t.Errorf("空切片不应写入缓存, Exists=%d", n)
	}
}
