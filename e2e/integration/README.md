# Bug 端到端测试

## 测试文件说明

### Go 测试

**文件**: `bug_lifecycle_test.go`

包含以下测试用例：

1. **TestBugLifecycle** - 完整的缺陷生命周期测试
   - 步骤1: 创建缺陷
   - 步骤2: 列表查询验证缺陷存在
   - 步骤3: 获取可用状态列表
   - 步骤4: 修改缺陷状态
   - 步骤5: 验证状态已更新
   - 步骤6: 删除缺陷
   - 步骤7: 验证缺陷已被删除

2. **TestBugCreateSimple** - 简单创建缺陷测试

3. **TestBugView** - 查看缺陷详情测试

4. **TestBugList** - 查询缺陷列表测试

5. **TestBugDelete** - 删除缺陷测试

### Shell 脚本测试

**文件**: `../../script/test_bug_lifecycle.sh`

Shell 脚本测试，与 Go 测试流程相同，但使用 bash 实现，适合 CI/CD 环境。

## 运行测试

### 环境变量配置

```bash
export LC_WORKSPACE_KEY="XXJSxiaobaice"    # 研发空间 key
export LC_TEST_PROJECT_ID="R24113J3C04"    # 测试项目 ID
```

### 运行 Go 测试

```bash
# 运行所有 Bug 测试
go test -v ./e2e/integration/... -run "TestBug"
```

### 运行 Shell 脚本测试

```bash
bash ./script/test_bug_lifecycle.sh
```

### 运行所有端到端测试

```bash
make test-e2e
```

## 注意事项

1. 测试需要有效的 `~/.lc/config.json` 配置文件
2. 测试会在真实环境中创建和删除缺陷，请确保使用测试项目
3. 测试会自动清理创建的缺陷，但如果测试中断可能需要手动清理
