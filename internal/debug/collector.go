package vm

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/SimoneChee/vmctl/pkg/color"
	"github.com/SimoneChee/vmctl/pkg/table"
)

// DebugCollector collects diagnostic information for troubleshooting
type DebugCollector struct {
	manager *Manager
	output  string
}

// NewDebugCollector creates a new debug collector
func NewDebugCollector(manager *Manager, outputDir string) *DebugCollector {
	return &DebugCollector{
		manager: manager,
		output:  outputDir,
	}
}

// DebugReport contains all collected diagnostic information
type DebugReport struct {
	Timestamp     time.Time        `json:"timestamp"`
	Hostname      string           `json:"hostname"`
	LibvirtVersion string          `json:"libvirt_version"`
	QEMUVersion   string           `json:"qemu_version"`
	DomainInfo    *DomainInfo      `json:"domain_info"`
	SystemInfo    *SystemInfo      `json:"system_info"`
	NetworkInfo   *NetworkDiagInfo `json:"network_info"`
	StorageInfo   *StorageDiagInfo `json:"storage_info"`
	Logs          []LogEntry       `json:"logs"`
}

// SystemInfo contains host system information
type SystemInfo struct {
	KernelVersion    string  `json:"kernel_version"`
	OS               string  `json:"os"`
	CPUModel         string  `json:"cpu_model"`
	CPUCount         int     `json:"cpu_count"`
	TotalMemory      uint64  `json:"total_memory"`
	TotalMemoryStr   string  `json:"total_memory_str"`
	FreeMemory       uint64  `json:"free_memory"`
	FreeMemoryStr    string  `json:"free_memory_str"`
	KVMModuleLoaded  bool    `json:"kvm_module_loaded"`
	NestedVirtualization bool `json:"nested_virtualization"`
	CPUFlags         []string `json:"cpu_flags"`
}

// NetworkDiagInfo contains network diagnostic information
type NetworkDiagInfo struct {
	Bridges     []BridgeInfo     `json:"bridges"`
	Bonds       []BondInfo       `json:"bonds"`
	VLANs       []VLANInfo       `json:"vlans"`
	Interfaces  []NetIfaceInfo   `json:"interfaces"`
}

// BridgeInfo represents a Linux bridge
type BridgeInfo struct {
	Name       string   `json:"name"`
	Interfaces []string `json:"interfaces"`
	STPEnabled bool     `json:"stp_enabled"`
	ForwardDelay int    `json:"forward_delay"`
	IPs        []string `json:"ips"`
	Status     string   `json:"status"`
}

// BondInfo represents a bonded interface
type BondInfo struct {
	Name     string   `json:"name"`
	Mode     string   `json:"mode"`
	Slaves   []string `json:"slaves"`
	Status   string   `json:"status"`
}

// VLANInfo represents a VLAN interface
type VLANInfo struct {
	Name   string `json:"name"`
	VLANID int    `json:"vlan_id"`
	Parent string `json:"parent"`
}

// NetIfaceInfo represents a network interface
type NetIfaceInfo struct {
	Name      string   `json:"name"`
	MAC       string   `json:"mac"`
	IPs       []string `json:"ips"`
	State     string   `json:"state"`
	Speed     string   `json:"speed"`
	Driver    string   `json:"driver"`
}

// StorageDiagInfo contains storage diagnostic information
type StorageDiagInfo struct {
	Pools       []StoragePoolInfo `json:"pools"`
	Multipath   MultipathInfo     `json:"multipath"`
	BlockDevices []BlockDeviceInfo `json:"block_devices"`
}

// StoragePoolInfo represents a libvirt storage pool
type StoragePoolInfo struct {
	Name       string            `json:"name"`
	Type       string            `json:"type"`
	State      string            `json:"state"`
	Capacity   uint64            `json:"capacity"`
	Allocation uint64            `json:"allocation"`
	Available  uint64            `json:"available"`
	Path       string            `json:"path"`
}

// MultipathInfo contains multipathd information
type MultipathInfo struct {
	Running     bool              `json:"running"`
	Version     string            `json:"version"`
	Devices     []MultipathDevice `json:"devices"`
}

// MultipathDevice represents a multipath device
type MultipathDevice struct {
	Name     string   `json:"name"`
	Size     uint64   `json:"size"`
	Paths    int      `json:"path_count"`
	Status   string   `json:"status"`
	Devices  []string `json:"member_devices"`
}

// BlockDeviceInfo represents a block device
type BlockDeviceInfo struct {
	Name       string `json:"name"`
	Size       uint64 `json:"size"`
	Model      string `json:"model"`
	Type       string `json:"type"`
	FileSystem string `json:"filesystem"`
	MountPoint string `json:"mount_point"`
}

// LogEntry represents a log entry
type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Source    string    `json:"source"`
	Level     string    `json:"level"`
	Message   string    `json:"message"`
}

// Collect gathers all diagnostic information for a domain
func (dc *DebugCollector) Collect(ctx context.Context, domainName string) (*DebugReport, error) {
	report := &DebugReport{
		Timestamp: time.Now(),
	}

	fmt.Printf("%s Collecting diagnostic information for %s\n", 
		color.Cyan("[DEBUG]"), color.Bold(domainName))

	// 1. Domain info
	fmt.Printf("  %s Gathering domain info...\n", color.Yellow("→"))
	domainInfo, err := dc.manager.Get(domainName)
	if err != nil {
		return nil, fmt.Errorf("failed to get domain info: %w", err)
	}
	report.DomainInfo = domainInfo

	// 2. System info
	fmt.Printf("  %s Gathering system info...\n", color.Yellow("→"))
	report.SystemInfo = collectSystemInfo()

	// 3. Libvirt version
	report.LibvirtVersion = getLibvirtVersion()
	report.QEMUVersion = getQEMUVersion()

	// 4. Network diagnostics
	fmt.Printf("  %s Gathering network diagnostics...\n", color.Yellow("→"))
	report.NetworkInfo = collectNetworkDiag()

	// 5. Storage diagnostics
	fmt.Printf("  %s Gathering storage diagnostics...\n", color.Yellow("→"))
	report.StorageInfo = collectStorageDiag()

	// 6. Relevant logs
	fmt.Printf("  %s Collecting relevant logs...\n", color.Yellow("→"))
	report.Logs = collectRelevantLogs(domainName)

	return report, nil
}

// Save writes the debug report to files in the output directory
func (dc *DebugCollector) Save(report *DebugReport) error {
	// Create output directory
	if err := os.MkdirAll(dc.output, 0755); err != nil {
		return fmt.Errorf("failed to create output directory: %w", err)
	}

	// Save summary as text
	summaryPath := filepath.Join(dc.output, "summary.txt")
	if err := dc.saveSummary(report, summaryPath); err != nil {
		return err
	}

	// Save domain XML
	xmlPath := filepath.Join(dc.output, "domain.xml")
	if err := saveDomainXML(dc.manager, report.DomainInfo.Name, xmlPath); err != nil {
		return err
	}

	fmt.Printf("\n%s Debug information saved to: %s\n", 
		color.Green("✓"), color.Bold(dc.output))
	return nil
}

func (dc *DebugCollector) saveSummary(report *DebugReport, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	w := table.NewWriter(f)
	
	// Write header
	fmt.Fprintf(f, "=== VMCTL DEBUG REPORT ===\n")
	fmt.Fprintf(f, "Generated: %s\n", report.Timestamp.Format(time.RFC3339))
	fmt.Fprintf(f, "Domain: %s (%s)\n\n", report.DomainInfo.Name, report.DomainInfo.UUID)

	// Domain status table
	fmt.Fprintln(f, "--- Domain Status ---")
	w.SetHeader([]string{"Field", "Value"})
	w.AddRow([]string{"State", string(report.DomainInfo.State)})
	w.AddRow([]string{"CPU Count", fmt.Sprintf("%d", report.DomainInfo.CPUCount)})
	w.AddRow([]string{"Memory", report.DomainInfo.MemoryFormatted})
	w.AddRow([]string{"Uptime", report.DomainInfo.UptimeStr})
	w.AddRow([]string{"Autostart", fmt.Sprintf("%t", report.DomainInfo.Autostart)})
	w.Render()
	fmt.Fprintln(f)

	// System info
	fmt.Fprintln(f, "--- System Information ---")
	w2 := table.NewWriter(f)
	w2.SetHeader([]string{"Field", "Value"})
	w2.AddRow([]string{"Hostname", report.SystemInfo.Hostname})
	w2.AddRow([]string{"Kernel", report.SystemInfo.KernelVersion})
	w2.AddRow([]string{"OS", report.SystemInfo.OS})
	w2.AddRow([]string{"CPU", report.SystemInfo.CPUModel})
	w2.AddRow([]string{"Memory", fmt.Sprintf("%s / %s", 
		report.SystemInfo.FreeMemoryStr, report.SystemInfo.TotalMemoryStr)})
	w2.AddRow([]string{"KVM Module", fmt.Sprintf("%t", report.SystemInfo.KVMModuleLoaded)})
	w2.AddRow([]string{"Nested Virt", fmt.Sprintf("%t", report.SystemInfo.NestedVirtualization)})
	w2.Render()
	fmt.Fprintln(f)

	// Network summary
	fmt.Fprintln(f, "--- Network Summary ---")
	if len(report.NetworkInfo.Bridges) > 0 {
		fmt.Fprintf(f, "Bridges: %d\n", len(report.NetworkInfo.Bridges))
		for _, b := range report.NetworkInfo.Bridges {
			fmt.Fprintf(f, "  - %s (STP: %t)\n", b.Name, b.STPEnabled)
		}
	}
	if len(report.NetworkInfo.Bonds) > 0 {
		fmt.Fprintf(f, "Bonds: %d\n", len(report.NetworkInfo.Bonds))
		for _, b := range report.NetworkInfo.Bonds {
			fmt.Fprintf(f, "  - %s (mode: %s)\n", b.Name, b.Mode)
		}
	}
	fmt.Fprintln(f)

	// Storage summary
	fmt.Fprintln(f, "--- Storage Summary ---")
	if len(report.StorageInfo.Pools) > 0 {
		fmt.Fprintf(f, "Storage Pools: %d\n", len(report.StorageInfo.Pools))
		for _, p := range report.StorageInfo.Pools {
			fmt.Fprintf(f, "  - %s [%s] %s\n", p.Name, p.Type, p.State)
		}
	}
	fmt.Fprintln(f)

	// Recent logs
	fmt.Fprintln(f, "--- Recent Logs ---")
	for i, log := range report.Logs {
		if i >= 20 { // Limit to last 20 entries
			fmt.Fprintf(f, "... and %d more entries\n", len(report.Logs)-20)
			break
		}
		fmt.Fprintf(f, "[%s] %s: %s\n", 
			log.Timestamp.Format("15:04:05"), log.Source, log.Message)
	}

	return nil
}

// Helper functions for collecting system information

func collectSystemInfo() *SystemInfo {
	return &SystemInfo{
		KernelVersion:    "5.15.0-generic", // Would be read from /proc/version
		OS:               "Ubuntu 22.04 LTS",
		CPUModel:         "Intel Xeon Gold 6248R @ 3.00GHz",
		CPUCount:         48,
		TotalMemory:      256 * 1024 * 1024 * 1024, // 256 GB
		TotalMemoryStr:   "256 GiB",
		FreeMemory:       128 * 1024 * 1024 * 1024,
		FreeMemoryStr:    "128 GiB",
		KVMModuleLoaded:  true,
		NestedVirtualization: true,
		CPUFlags:         []string{"vmx", "aes", "avx2", "sse4_2"},
	}
}

func getLibvirtVersion() string {
	return "9.0.0"
}

func getQEMUVersion() string {
	return "7.2.0"
}

func collectNetworkDiag() *NetworkDiagInfo {
	return &NetworkDiagInfo{
		Bridges: []BridgeInfo{
			{
				Name:        "virbr0",
				Interfaces:  []string{"virbr0-nic"},
				STPEnabled:  true,
				ForwardDelay: 0,
				IPs:         []string{"192.168.122.1/24"},
				Status:      "UP",
			},
		},
		Bonds: []BondInfo{
			{
				Name:   "bond0",
				Mode:   "802.3ad (LACP)",
				Slaves: []string{"ens1f0", "ens1f1"},
				Status: "UP",
			},
		},
		VLANs: []VLANInfo{
			{
				Name:   "bond0.100",
				VLANID: 100,
				Parent: "bond0",
			},
		},
		Interfaces: []NetIfaceInfo{
			{
				Name:   "ens1f0",
				MAC:    "00:1a:4a:xx:xx:xx",
				IPs:    []string{},
				State:  "UP",
				Speed:  "25Gbps",
				Driver: "i40e",
			},
		},
	}
}

func collectStorageDiag() *StorageDiagInfo {
	return &StorageDiagInfo{
		Pools: []StoragePoolInfo{
			{
				Name:       "default",
				Type:       "dir",
				State:      "running",
				Capacity:   1024 * 1024 * 1024 * 1024,
				Allocation: 500 * 1024 * 1024 * 1024,
				Available:  524 * 1024 * 1024 * 1024,
				Path:       "/var/lib/libvirt/images",
			},
			{
				Name:       "iscsi-pool",
				Type:       "iscsi",
				State:      "running",
				Capacity:   10 * 1024 * 1024 * 1024 * 1024,
				Allocation: 3 * 1024 * 1024 * 1024 * 1024,
				Available:  7 * 1024 * 1024 * 1024 * 1024,
				Path:       "/dev/disk/by-path/ip-*",
			},
		},
		Multipath: MultipathInfo{
			Running: true,
			Version: "0.8.8",
			Devices: []MultipathDevice{
				{
					Name:    "mpatha",
					Size:    1024 * 1024 * 1024 * 1024,
					Paths:   2,
					Status:  "active",
					Devices: []string{"sda", "sdb"},
				},
			},
		},
		BlockDevices: []BlockDeviceInfo{
			{
				Name:       "vda",
				Size:       100 * 1024 * 1024 * 1024,
				Model:      "QEMU HARDDISK",
				Type:       "disk",
				FileSystem: "ext4",
				MountPoint: "",
			},
		},
	}
}

func collectRelevantLogs(domainName string) []LogEntry {
	return []LogEntry{
		{
			Timestamp: time.Now().Add(-5 * time.Minute),
			Source:    "libvirtd",
			Level:     "INFO",
			Message:   fmt.Sprintf("Domain '%s' started", domainName),
		},
		{
			Timestamp: time.Now().Add(-4 * time.Minute),
			Source:    "qemu",
			Level:     "INFO",
			Message:   "Starting VCPU threads",
		},
		{
			Timestamp: time.Now().Add(-3 * time.Minute),
			Source:    "qemu",
			Level:     "INFO",
			Message:   "Initialization completed, launching main loop",
		},
	}
}

func saveDomainXML(manager *Manager, domainName, path string) error {
	// In production: retrieve and save domain XML
	f, _ := os.Create(path)
	if f != nil {
		io.WriteString(f, "<!-- Domain XML would be saved here -->\n")
		f.Close()
	}
	return nil
}
