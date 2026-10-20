// Package main is the entry point for vmctl - a modern CLI for KVM/Libvirt VM management.
//
// vmctl provides a user-friendly interface to libvirt, inspired by kubectl and docker CLI design.
package main

import (
	"os"

	"github.com/SimoneChee/vmctl/internal/config"
	"github.com/SimoneChee/vmctl/pkg/color"
	"github.com/spf13/cobra"
)

var (
	cfgFile string
	output  string
	noColor bool
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "vmctl",
	Short: "A modern CLI for KVM/Libvirt virtual machine management",
	Long: `vmctl ☁️ - A modern CLI for KVM/Libvirt virtual machine management.

Like kubectl, but for VMs. Manage your virtual machines with ease using
a beautiful, intuitive command-line interface built on top of libvirt.

Examples:
  # List all running VMs
  vmctl list

  # Get detailed info about a VM
  vmctl info my-vm

  # Create a new VM
  vmctl create my-vm --cpu 4 --memory 8G --disk 100G

  # Start/stop/reboot VMs
  vmctl start my-vm
  vmctl stop my-vm --graceful

  # Real-time monitoring (like top for VMs)
  vmctl monitor

  # Collect debug information
  vmctl debug my-vm --output ./debug-dir`,
	Version: "v0.1.0-alpha",
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "Config file (default is $HOME/.vmctl.yaml)")
	rootCmd.PersistentFlags().StringVarP(&output, "output", "o", "table", "Output format: table|json|yaml|wide")
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")
}

func main() {
	// Disable color if requested or if output is not a terminal
	if noColor || !isTerminal() {
		color.Disable()
	}

	if err := rootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}

// isTerminal checks if stdout is a terminal
func isTerminal() bool {
	fi, _ := os.Stdout.Stat()
	return (fi.Mode() & os.ModeCharDevice) != 0
}

func init() {
	cobra.OnInitialize(initConfig)
}

func initConfig() {
	config.InitConfig(cfgFile)
}
