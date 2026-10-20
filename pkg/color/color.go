// Package color provides terminal color support for vmctl output
package color

import (
	"fmt"

	"github.com/fatih/color"
)

var (
	// Predefined colors for consistent theming
	Header    = color.New(color.Bold, color.FgHiWhite)
	Success   = color.New(color.FgGreen)
	Warning   = color.New(color.FgYellow)
	Error     = color.New(color.FgRed)
	Info      = color.New(color.FgCyan)
	Highlight = color.New(color.Bold, color.FgHiMagenta)
	Dim       = color.New(color.FgHiBlack)
	Muted     = color.New(color.FgHiBlack)

	// State colors
	Running  = color.New(color.FgGreen, color.Bold)
	Stopped  = color.New(color.FgRed)
	Paused   = color.New(color.FgYellow)
	Unknown  = color.New(color.FgHiBlack)

	// Resource colors
	CPUUsage    = color.New(color.FgBlue)
	MemoryUsage = color.New(color.FgMagenta)
	DiskUsage   = color.New(color.FgCyan)
	NetworkIO   = color.New(color.FgHiGreen)

	// Table header
	TableHeader = color.New(color.Bold, color.BgHiBlack, color.FgWhite)
)

// StateColor returns the appropriate color for a domain state
func StateColor(state string) *color.Color {
	switch state {
	case "running":
		return Running
	case "stopped", "shut off":
		return Stopped
	case "paused":
		return Paused
	default:
		return Unknown
	}
}

// UsageColor returns color based on usage percentage (0-100)
func UsageColor(percentage float64) *color.Color {
	switch {
	case percentage >= 90:
		return Error
	case percentage >= 70:
		return Warning
	default:
		return Success
	}
}

// PrintBanner prints the vmctl banner
func PrintBanner() {
	banner := `
   ____ _                 ___                
  / ___| | ___  _ __ __ _| |_ _  ___ _ _ __  
 | |  _ |/ _ \| '__/ _` + "`" + ` | __| |/ _ \ | '_ \ 
 | |_| | (_) | | | (_| | |_| |  __/ | | | |
  \____|\___/|_|  \__,_|\__|_|\___|_|_| |_|`
	
	fmt.Println(Header.Sprint(banner))
	fmt.Println(Info.Sprint("  A modern CLI for KVM/Libvirt VM management"))
	fmt.Println()
}

// Helper functions that wrap color functions for convenience

func Cyan(format string, a ...interface{}) string { return color.Cyan.Sprintf(format, a...) }
func Green(format string, a ...interface{}) string { return color.Green.Sprintf(format, a...) }
func Red(format string, a ...interface{}) string { return color.Red.Sprintf(format, a...) }
func Yellow(format string, a ...interface{}) string { return color.Yellow.Sprintf(format, a...) }
func Bold(format string, a ...interface{}) string { return color.New(color.Bold).Sprintf(format, a...) }
func HiWhite(format string, a ...interface{}) string { return color.HiWhite.Sprintf(format, a...) }
func HiBlack(format string, a ...interface{}) string { return color.HiBlack.Sprintf(format, a...) }
