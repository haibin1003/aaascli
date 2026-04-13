# Cache Package

通用缓存层实现，提供线程安全、支持 TTL 的高性能缓存。

## 特性

- **泛型支持**: 支持任意 key-value 类型
- **线程安全**: 使用读写锁保证并发安全
- **TTL 支持**: 支持自定义过期时间
- **自动清理**: 支持后台自动清理过期项
- **多缓存管理**: 内置 Manager 管理多种类型的缓存

## 使用示例

### 基础缓存

```go
// 创建缓存
cache := cache.New[string, string](5 * time.Minute)

// 设置值
cache.Set("key", "value")

// 获取值
value, found := cache.Get("key")
if found {
    fmt.Println(value)
}

// 自定义 TTL
cache.SetWithTTL("temp", "data", 30 * time.Second)

// 清理过期项
cache.Cleanup()

// 启动自动清理
stop := cache.AutoCleanup(1 * time.Minute)
defer stop()
```

### 缓存管理器

```go
// 创建管理器
manager := cache.NewManager()

// 工作空间缓存
manager.SetWorkspace("XXJSLJCLIDEV", "object-id-123")
objectID, found := manager.GetWorkspace("XXJSLJCLIDEV")

// 用户信息缓存
manager.SetUser("user-123", cache.UserInfo{
    UserID: "123",
    Username: "john",
})

// API 响应缓存
manager.SetAPIResponse("/api/users", jsonData)

// 查看统计
stats := manager.Stats()
fmt.Printf("Workspace cache: %d items\n", stats["workspace"])
```

### GetOrSet 模式

```go
value, err := cache.GetOrSet("key", func() (string, error) {
    // 只在缓存未命中时执行
    return fetchFromDatabase()
})
```

## 默认 TTL

| 缓存类型 | 默认 TTL |
|---------|---------|
| Workspace | 5 分钟 |
| User Info | 10 分钟 |
| Config | 1 小时 |
| API Response | 30 秒 |

## 性能

```
BenchmarkCache_Get-8           100000000    11.2 ns/op
BenchmarkCache_Set-8           50000000     28.4 ns/op
BenchmarkCache_ConcurrentGet-8  20000000    62.1 ns/op
```

## 测试

```bash
go test ./internal/cache/... -v
```
