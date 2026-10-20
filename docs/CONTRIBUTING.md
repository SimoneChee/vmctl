# Contributing to vmctl

感谢你考虑为 vmctl 贡献代码！以下是一些约定，帮助你顺利提交。

## 快速开始

1. **Fork** 本仓库到你的账号
2. 克隆你的 fork：`git clone https://github.com/<your-name>/vmctl.git`
3. 创建功能分支：`git checkout -b feat/your-feature`
4. 开发并提交
5. 推送到你的 fork 并发起 **Pull Request**

## 开发环境

- Go 1.21+
- 建议在 Linux 环境开发（依赖 libvirt）
- 安装 libvirt 开发库：`sudo apt install libvirt-dev`

## 目录约定

```
cmd/vmctl/           # CLI 入口
internal/vm/         # 虚拟机管理核心
internal/debug/      # 诊断收集
internal/config/     # 配置管理
pkg/color/           # 颜色主题
pkg/table/           # 表格渲染
```

## 代码规范

- 遵循 `gofmt` 和 `go vet`
- 导出函数/类型需写注释
- 错误信息使用 `fmt.Errorf` 包装，保留原始错误上下文

```go
// ✅ 推荐
if err != nil {
    return fmt.Errorf("failed to start domain '%s': %w", name, err)
}

// ❌ 不推荐
if err != nil {
    return err
}
```

## 提交信息

使用约定式提交（Conventional Commits）：

- `feat:` 新功能
- `fix:` 修复
- `docs:` 文档
- `refactor:` 重构
- `test:` 测试

示例：`feat: add snapshot list command`

## 测试

```bash
go test ./... -v
go test ./... -cover
```

## 报告 Bug

提交 Issue 时请包含：
1. 操作系统与 libvirt 版本
2. 复现步骤
3. 实际输出与预期输出
4. `vmctl debug` 收集的诊断信息（如适用）

## 行为准则

请保持友善、专业。我们欢迎所有水平的贡献者。
