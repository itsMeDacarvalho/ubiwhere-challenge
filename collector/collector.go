/*
	Author	:	Daniel Alexandre Neves de Carvalho
	Date	:	15/06/2020
	File	:	collector.go
	Overview: 	Collector only provides two functions to get OS data.
				CPU and RAM are collected from here.
*/

package collector

import (
	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// GetRAM - Collect RAM info from OS
func GetRAM() []uint64 {
	cpuData := make([]uint64, 2)
	memStats, _ := mem.VirtualMemory()

	cpuData[0] = memStats.Total / (1024 * 1024)
	cpuData[1] = memStats.Used / (1024 * 1024)

	return cpuData
}

// GetCPU - Collect CPU info from OS
func GetCPU() float64 {
	cpuStats, _ := cpu.Percent(0, true)

	return cpuStats[0]
}
