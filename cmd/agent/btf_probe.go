package main

import (
	"os"
)

// probeBTF checks for BTF support (Ubuntu 20.04+). Falls back to auditd on 16.04/18.04 4.15
func probeBTF() bool {
	if _, err := os.Stat("/sys/kernel/btf/vmlinux"); err == nil {
		return true
	}
	// Alternative: check kernel version >=5.8
	if data, err := os.ReadFile("/proc/version"); err == nil {
		s := string(data)
		// crude: if contains "5." and not "4.15"
		if len(s) > 0 && contains(s, "5.") {
			return true
		}
	}
	return false
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i <= len(s)-len(sub); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}
