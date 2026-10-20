# vmctl Roadmap

> 这个文件记录 vmctl 的演进计划。欢迎提 Issue 或 PR 一起完善。

## 版本规划

### v0.1.0-alpha（当前）

基础能力，打通 libvirt 的核心操作链路。

- [x] `list` — 列出所有虚拟机（含状态、CPU、内存）
- [x] `info` — 查看虚拟机详细信息
- [x] `start` / `stop` / `reboot` — 生命周期管理
- [x] `debug` — 一键收集诊断信息
- [x] 彩色表格输出 + JSON 输出格式

### v0.2.0

补全创建与快照能力，完善用户体验。

- [ ] `create` — 交互式/命令行快速创建虚拟机
- [ ] `snapshot create/list/restore` — 快照管理
- [ ] `destroy` — 强制销毁（含确认交互）
- [ ] Shell 自动补全（bash / zsh / fish）
- [ ] `--wide` 输出模式（显示 IP、MAC、uptime）

### v0.3.0

实时监控与迁移。

- [ ] `monitor` — 类 `htop` 的实时资源监控（CPU/内存/网络/磁盘 IO）
- [ ] `migrate` — 实时迁移（live migration）带进度条
- [ ] 多 libvirt URI 管理（本机 + 远程 SSH）
- [ ] `vmctl top` — 宿主机维度的资源总览

### v1.0.0

稳定版目标。

- [ ] 完整的单元测试与集成测试（>80% 覆盖率）
- [ ] 配置文件 `.vmctl.yaml` 全量支持（别名、默认值、主题）
- [ ] 多语言输出（中英）
- [ ] 性能优化：连接池、并发查询
- [ ] 官方 Release 二进制（Linux / macOS / Windows）

## 远期探索

- [ ] 插件系统（自定义子命令）
- [ ] TUI 交互模式（如 `lazydocker` 风格）
- [ ] Web UI / 只读仪表盘
- [ ] Terraform provider 集成

---

## 版本节奏

| 版本 | 目标 | 状态 |
|------|------|------|
| v0.1.0-alpha | 基础生命周期 | 🚧 开发中 |
| v0.2.0 | 创建 + 快照 | 📝 规划 |
| v0.3.0 | 监控 + 迁移 | 📝 规划 |
| v1.0.0 | 稳定版 | 📝 规划 |

> 时间线取决于社区反馈与实际需求，欢迎在 Issue 里投票你最想要的功能。
