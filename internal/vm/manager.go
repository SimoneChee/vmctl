// Package vm provides virtual machine management functionality for vmctl.
package vm

import (
	"context"
	"fmt"
	"time"

	"github.com/libvirt/libvirt-go"
)

// DomainState represents the state of a virtual machine
type DomainState string

const (
	DomainStateRunning  DomainState = "running"
	DomainStateStopped  DomainState = "stopped"
	DomainStatePaused   DomainState = "paused"
	DomainStateShutdown DomainState = "shutdown"
	DomainStateCrashed  DomainState = "crashed"
	DomainStateUnknown  DomainState = "unknown"
)

// DomainInfo contains comprehensive information about a virtual machine
type DomainInfo struct {
	Name          string            `json:"name"`
	UUID          string            `json:"uuid"`
	State         DomainState       `json:"state"`
	CPUCount      uint              `json:"cpu_count"`
	MemoryMB      uint64            `json:"memory_mb"`
	MemoryFormatted string          `json:"memory_formatted"`
	DiskSizeGB    float64           `json:"disk_size_gb"`
	VCPUUsage     float64           `json:"vcpu_usage"`
	MemoryUsage   float64           `json:"memory_usage"`
	IPAddresses   []string          `json:"ip_addresses"`
	MACAddresses  []string          `json:"mac_addresses"`
	OSType        string            `json:"os_type"`
	Architecture  string            `json:"architecture"`
	Uptime        time.Duration     `json:"uptime"`
	UptimeStr     string            `json:"uptime_str"`
	Autostart     bool              `json:"autostart"`
	Persistent    bool              `json:"persistent"`
	Disks         []DiskInfo        `json:"disks"`
	Interfaces    []InterfaceInfo   `json:"interfaces"`
	Snapshots     []SnapshotInfo    `json:"snapshots"`
}

// DiskInfo represents a disk attached to a VM
type DiskInfo struct {
	Target   string  `json:"target"`   // e.g., vda, sda
	Source   string  `json:"source"`   // image path or block device
	Capacity uint64  `json:"capacity"` // bytes
	Allocation uint64 `json:"allocation"` // actual allocation
	Physical  uint64  `json:"physical"`  // physical size
	Driver   string  `json:"driver"`   // qcow2, raw, etc.
	Format   string  `json:"format"`
}

// InterfaceInfo represents a network interface of a VM
type InterfaceInfo struct {
	MAC      string   `json:"mac"`
	Network  string   `json:"network"`  // network name
	Model    string   `json:"model"`    // virtio, e1000, etc.
	IPs      []string `json:"ips"`
	State    string   `json:"state"`
}

// SnapshotInfo represents a domain snapshot
type SnapshotInfo struct {
	Name        string    `json:"name"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
	Size        uint64    `json:"size"`
	State       string    `json:"state"`
	Parent      string    `json:"parent"`
}

// Manager provides high-level VM management operations
type Manager struct {
	conn *libvirt.Connect
}

// NewManager creates a new VM manager with libvirt connection
func NewManager(uri string) (*Manager, error) {
	conn, err := libvirt.NewConnect(uri)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to libvirt: %w", err)
	}
	return &Manager{conn: conn}, nil
}

// NewManagerFromDefault connects to the default libvirt instance
func NewManagerFromDefault() (*Manager, error) {
	return NewManager("qemu:///system")
}

// Close closes the libvirt connection
func (m *Manager) Close() error {
	if m.conn != nil {
		_, err := m.conn.Close()
		return err
	}
	return nil
}

// List returns all domains (VMs) with their basic info
func (m *Manager) List(all bool) ([]DomainInfo, error) {
	flags := libvirt.CONNECT_LIST_DOMAINS_ACTIVE
	if all {
		flags |= libvirt.CONNECT_LIST_DOMAINS_INACTIVE
	}

	doms, err := m.conn.ListAllDomains(flags)
	if err != nil {
		return nil, fmt.Errorf("failed to list domains: %w", err)
	}

	var infos []DomainInfo
	for _, dom := range doms {
		info, mErr := m.getDomainInfo(dom)
		if mErr != nil {
			// Skip domains that fail to query but continue listing others
			continue
		}
		infos = append(infos, info)
		dom.Free()
	}

	return infos, nil
}

// Get returns detailed information about a specific domain by name
func (m *Manager) Get(name string) (*DomainInfo, error) {
	dom, err := m.conn.LookupDomainByName(name)
	if err != nil {
		return nil, fmt.Errorf("domain '%s' not found: %w", name, err)
	}
	defer dom.Free()

	info, err := m.getDomainInfo(dom)
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// getDomainInfo extracts comprehensive info from a libvirt domain
func (m *Manager) getDomainInfo(dom *libvirt.Domain) (DomainInfo, error) {
	var info DomainInfo

	// Basic info
	name, _ := dom.GetName()
	info.Name = name

	uuid, _ := dom.GetUUIDString()
	info.UUID = uuid

	state, _, _ := dom.GetState()
	info.State = libvirtStateToDomainState(state)

	// CPU and Memory
	info.CPUCount, _ = dom.GetVcpusFlags(libvirt.DOMAIN_VCPU_CURRENT)
	info.MemoryMB, _ = dom.GetMemoryStats(0, 1, 0) // in KB
	info.MemoryMB /= 1024 // convert to MB
	info.MemoryFormatted = formatBytes(info.MemoryMB * 1024 * 1024)

	// Domain info (max memory, CPUs, etc.)
	domInfo, err := dom.GetInfo()
	if err == nil {
		if info.MemoryMB == 0 {
			info.MemoryMB = domInfo.Memory / 1024
			info.MemoryFormatted = formatBytes(domInfo.Memory * 1024)
		}
		if info.CPUCount == 0 {
			info.CPUCount = uint(domInfo.NrVirtCpu)
		}
	}

	// Autostart
	info.Autostart, _ = dom.GetAutostart()
	info.Persistent, _ = dom.IsPersistent()

	// OS type and architecture
	osType, _ := dom.GetOSType()
	info.OSType = osType
	xmlDesc, _ := dom.GetXMLDesc(0)
	info.Architecture = extractArchFromXML(xmlDesc)

	// Uptime
	if info.State == DomainStateRunning {
		info.Uptime, _ = dom.GetInfo()
		// Calculate actual uptime from creation time
		createTime, err := dom.GetCreationTime()
		if err == nil {
			info.Uptime = time.Since(time.Unix(int64(createTime), 0))
		}
		info.UptimeStr = formatDuration(info.Uptime)
	}

	// Disk information
	info.Disks, _ = m.getDiskInfo(dom, xmlDesc)

	// Network interface information
	info.Interfaces, _ = m.getInterfaceInfo(dom, xmlDesc)

	// Collect IP addresses from interfaces
	for _, iface := range info.Interfaces {
		info.IPAddresses = append(info.IPAddresses, iface.IPs...)
		info.MACAddresses = append(info.MACAddresses, iface.MAC)
	}

	// Snapshots
	snapshots, _ := m.getSnapshotList(dom)
	info.Snapshots = snapshots

	// Total disk size
	for _, disk := range info.Disks {
		info.DiskSizeGB += float64(disk.Capacity) / (1024 * 1024 * 1024)
	}

	return info, nil
}

// Start starts a domain
func (m *Manager) Start(ctx context.Context, name string) error {
	dom, err := m.conn.LookupDomainByName(name)
	if err != nil {
		return fmt.Errorf("domain '%s' not found: %w", name, err)
	}
	defer dom.Free()

	if err := dom.Create(); err != nil {
		return fmt.Errorf("failed to start domain '%s': %w", name, err)
	}
	return nil
}

// Stop gracefully shuts down a domain (ACPI shutdown)
func (m *Manager) Stop(ctx context.Context, name string, graceful bool, timeout time.Duration) error {
	dom, err := m.conn.LookupDomainByName(name)
	if err != nil {
		return fmt.Errorf("domain '%s' not found: %w", name, err)
	}
	defer dom.Free()

	if graceful {
		if err := dom.Shutdown(); err != nil {
			return fmt.Errorf("failed to shutdown domain '%s': %w", name, err)
		}
		// Wait for shutdown with timeout
		done := make(chan error, 1)
		go func() {
			for {
				state, _, err := dom.GetState()
				if err != nil || state == libvirt.DOMAIN_SHUTOFF {
					done <- err
					return
				}
				time.Sleep(500 * time.Millisecond)
			}
		}()
		select {
		case <-ctx.Done():
			return ctx.Err()
		case err := <-done:
			return err
		case <-time.After(timeout):
			return fmt.Errorf("timeout waiting for graceful shutdown")
		}
	} else {
		if err := dom.Destroy(); err != nil {
			return fmt.Errorf("failed to force stop domain '%s': %w", name, err)
		}
	}
	return nil
}

// Reboot reboots a domain
func (m *Manager) Reboot(ctx context.Context, name string, graceful bool) error {
	dom, err := m.conn.LookupDomainByName(name)
	if err != nil {
		return fmt.Errorf("domain '%s' not found: %w", name, err)
	}
	defer dom.Free()

	if graceful {
		err = dom.Reboot(0)
	} else {
		err = dom.Reset(0)
	}
	if err != nil {
		return fmt.Errorf("failed to reboot domain '%s': %w", name, err)
	}
	return nil
}

// Migrate performs live migration of a domain
func (m *Manager) Migrate(ctx context.Context, name string, destURI string, live bool, 
	unsafe bool, progress chan<- float64) error {
	
	dom, err := m.conn.LookupDomainByName(name)
	if err != nil {
		return fmt.Errorf("domain '%s' not found: %w", name, err)
	}
	defer dom.Free()

	destConn, err := libvirt.NewConnect(destURI)
	if err != nil {
		return fmt.Errorf("failed to connect to destination: %w", err)
	}
	defer destConn.Close()

	flags := libvirt.MIGRATE_LIVE
	if unsafe {
		flags |= libvirt.MIGRATE_UNSAFE
	}

	// Perform migration
	newDom, err := dom.Migrate(destConn, flags, "", "", 0)
	if err != nil {
		return fmt.Errorf("migration failed: %w", err)
	}
	
	if newDom != nil {
		newDom.Free()
	}
	return nil
}

// Helper functions

func libvirtStateToDomainState(state libvirt.DomainState) DomainState {
	switch state {
	case libvirt.DOMAIN_RUNNING:
		return DomainStateRunning
	case libvirt.DOMAIN_BLOCKED:
		return DomainStatePaused
	case libvirt.DOMAIN_PAUSED:
		return DomainStatePaused
	case libvirt.DOMAIN_SHUTDOWN:
		return DomainStateShutdown
	case libvirt.DOMAIN_SHUTOFF:
		return DomainStateStopped
	case libvirt.DOMAIN_CRASHED:
		return DomainStateCrashed
	default:
		return DomainStateUnknown
	}
}

func formatBytes(bytes uint64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

func formatDuration(d time.Duration) string {
	if d < time.Second {
		return d.String()
	}
	if d < time.Minute {
		return fmt.Sprintf("%.0fs", d.Seconds())
	}
	if d < time.Hour {
		return fmt.Sprintf("%dm%ds", int(d.Minutes()), int(d.Seconds())%60)
	}
	return fmt.Sprintf("%dh%dm", int(d.Hours()), int(d.Minutes())%60)
}

func extractArchFromXML(xml string) string {
	// Simple XML parsing for architecture - in production use proper XML parser
	// This is a simplified version for demonstration
	if len(xml) > 0 {
		return "x86_64" // Default assumption
	}
	return "unknown"
}

// The following methods would be implemented with full XML parsing
// Using libvirt's XML description APIs

func (m *Manager) getDiskInfo(dom *libvirt.Domain, xmlDesc string) ([]DiskInfo, error) {
	// In production: parse XMLDesc to extract disk information
	// For now, return empty slice as placeholder
	return []DiskInfo{}, nil
}

func (m *Manager) getInterfaceInfo(dom *libvirt.Domain, xmlDesc string) ([]InterfaceInfo, error) {
	// In production: parse XMLDesc and query guest network info
	// For now, return empty slice as placeholder
	return []InterfaceInfo{}, nil
}

func (m *Manager) getSnapshotList(dom *libvirt.Domain) ([]SnapshotInfo, error) {
	// Query snapshot list from domain
	snapshots, err := dom.ListAllSnapshots(0)
	if err != nil {
		return nil, err
	}

	var result []SnapshotInfo
	for _, snap := range snapshots {
		var si SnapshotInfo
		si.Name, _ = snap.GetName()
		si.Description, _ = snap.GetDescription()
		si.CreatedAt, _ = snap.GetCreateTime()
		result = append(result, si)
		snap.Free()
	}
	return result, nil
}
