# vmctl ☁️

<p align="center">
  <strong>A modern CLI for KVM/Libvirt virtual machine management</strong><br>
  <em>Like kubectl, but for VMs</em>
</p>

<p align="center">
  <a href="https://github.com/SimoneChee/vmctl/actions"><img src="https://img.shields.io/github/actions/workflow/status/SimoneChee/vmctl/go.yml?branch=main" alt="Build Status"></a>
  <a href="https://goreportcard.com/report/github.com/SimoneChee/vmctl"><img src="https://goreportcard.com/badge/github.com/SimoneChee/vmctl" alt="Go Report Card"></a>
  <a href="https://github.com/SimoneChee/vmctl/blob/main/LICENSE"><img src="https://img.shields.io/github/license/SimoneChee/vmctl" alt="License"></a>
  <a href="https://github.com/SimoneChee/vmctl/stargazers"><img src="https://img.shields.io/github/stars/SimoneChee/vmctl?style=social" alt="Stars"></a>
</p>

---

## Why vmctl?

Managing KVM virtual machines with `virsh` is powerful but painful:

- **Too many flags**: `virsh dominfo --domain uuid` vs `vmctl info my-vm`
- **No color output**: Everything looks the same
- **No tab completion**: You have to remember everything
- **Scattered commands**: Domain info, network info, storage info are separate

**vmctl** solves all of this with a modern, user-friendly CLI inspired by `kubectl` and `docker`.

## Features

**Currently working (v0.1.x):**

- 🎨 **Color output**: State-colored tables for `list` / `info`
- 🔍 **Rich info**: VM state, CPU, memory, uptime, autostart in one command
- ⚡ **Lifecycle**: `start` / `stop` (graceful or force) / `reboot`
- 🐛 **Debug collector**: Gather system, network, storage, logs into a report

**Planned (see [ROADMAP.md](docs/ROADMAP.md)):**

- 📊 Monitoring mode (like `top` for VMs)
- 💾 Snapshot management
- 🔄 Live migration with progress tracking
- 📝 Tab completion (bash/zsh/fish)

## Quick Start

### Installation

> ⚠️ v0.1.x is an early alpha. The `create/snapshot/monitor/migrate` commands below are part of the design goal and are being implemented step by step — see [ROADMAP.md](docs/ROADMAP.md) for the current progress.

```bash
# Build from source (requires libvirt-dev & pkg-config)
sudo apt install libvirt-dev pkg-config
go install github.com/SimoneChee/vmctl/cmd/vmctl@latest
```

### Prerequisites

- Libvirt installed and running
- Go 1.21+ (if building from source)
- `libvirt-dev` and `pkg-config` (CGO dependency)

### Basic Usage

```bash
# List all VMs with status
$ vmctl list
 NAME          STATE    CPU    MEMORY    DISK     IP ADDRESS
───────────── ─────── ───── ──────── ─────── ──────────────
 web-server-01  running   4     8.0 GiB   100 GiB  192.168.1.10
 db-primary    running   8     32.0 GiB  500 GiB  192.168.1.20
 cache-node    stopped   2     4.0 GiB    50 GiB  -
 backup-vm     paused    2     8.0 GiB   200 GiB  192.168.1.30

# Get detailed info about a VM
$ vmctl info web-server-01

# Create a new VM quickly
$ vmctl create app-server \
  --cpu 4 --memory 8G --disk 100G \
  --network bridge=virbr0 \
  --image /var/lib/libvirt/images/ubuntu-22.04.qcow2

# Lifecycle management
$ vmctl start web-server-01
$ vmctl stop db-primary --graceful
$ vmctl force-stop stuck-vm
$ vmctl reboot cache-node

# Snapshot operations
$ vmctl snapshot list web-server-01
$ vmctl snapshot create web-server-01 --name pre-upgrade
$ vmctl snapshot restore web-server-01 --snapshot pre-upgrade

# Real-time monitoring (like htop for VMs)
$ vmctl monitor

# Debug: collect all diagnostic info
$ vmctl debug web-server-01 --output ./debug-dir

# Migration with progress
$ vmctl migrate web-server-01 --dest host2.example.com --live
```

## Commands Reference

| Command | Status | Description |
|---------|--------|-------------|
| `list` | ✅ v0.1 | List all domains with colored table |
| `info` | ✅ v0.1 | Show domain details (CPU/memory/uptime/autostart) |
| `start` | ✅ v0.1 | Start a domain |
| `stop` | ✅ v0.1 | Stop a domain (graceful with timeout, or force) |
| `reboot` | ✅ v0.1 | Reboot a domain (graceful or reset) |
| `debug` | ✅ v0.1 | Collect diagnostics into a report |
| `create` | 🚧 v0.2 | Create a new VM |
| `snapshot` | 🚧 v0.2 | Snapshot management |
| `migrate` | 🚧 v0.3 | Live migration |
| `monitor` | 🚧 v0.3 | Real-time monitoring |
| `completion` | 🚧 v0.2 | Shell completion |

## Architecture

```
┌─────────────────────────────────────────────────────────┐
│                        vmctl                             │
├─────────────────────────────────────────────────────────┤
│  cmd/vmctl/main.go                                       │
│  ┌─────────┐ ┌─────────┐ ┌──────────┐ ┌──────────────┐ │
│  │ vm cmd  │ │network  │ │ storage  │ │   debug      │ │
│  │         │ │  cmd    │ │   cmd    │ │    cmd       │ │
│  └────┬────┘ └────┬────┘ └────┬─────┘ └──────┬───────┘ │
│       │           │           │               │         │
│  ┌────▼───────────▼───────────▼───────────────▼───────┐  │
│  │                  internal/                         │  │
│  │  ┌──────┐ ┌───────┐ ┌─────────┐ ┌─────────────┐  │  │
│  │  │  vm  │ │network│ │ storage │ │    config    │  │  │
│  │  │mgr   │ │  mgr  │ │  mgr    │ │             │  │  │
│  │  └──────┘ └───────┘ └─────────┘ └─────────────┘  │  │
│  └────────────────────┬──────────────────────────────┘  │
│                       │                                 │
│  ┌────────────────────▼──────────────────────────────┐  │
│  │                    pkg/                            │  │
│  │  ┌─────────┐ ┌─────────┐ ┌─────────────────────┐  │  │
│  │  │ output  │ │  table  ｜ │  color              │  │  │
│  │  │(format) │ │(render) │ │  (theme)            │  │  │
│  │  └─────────┘ └─────────┘ └─────────────────────┘  │  │
│  └──────────────────────────────────────────────────┘  │
├─────────────────────────────────────────────────────────┤
│                 libvirt (libvirt-go)                     │
└─────────────────────────────────────────────────────────┘
```

## Project Status

This is an active project under development. Current version: **v0.1.0-alpha**

See [ROADMAP.md](docs/ROADMAP.md) for planned features.

## Contributing

Contributions are welcome! Please read [CONTRIBUTING.md](docs/CONTRIBUTING.md) first.

## License

[Apache License 2.0](LICENSE)

## Acknowledgments

- Inspired by [kubectl](https://kubernetes.io/docs/reference/kubectl/) UX design
- Built on [libvirt-go](https://github.com/libvirt/libvirt-go) bindings
- Table rendering inspired by [tablewriter](https://github.com/olekukonez/tablewriter)


- GitHub: [@SimoneChee](https://github.com/SimoneChee)
- Focus: KVM virtualization, Linux networking, distributed storage

---

<p align="center">
  <strong>If you find this useful, please give it a star ⭐</strong>
</p>
