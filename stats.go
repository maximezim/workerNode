package main

import (
	"fmt"
	"time"

	"github.com/shirou/gopsutil/cpu"
	"github.com/shirou/gopsutil/mem"
)

// getStats renvoie l'utilisation du CPU (%) et de la RAM (%)
func getStats() (WorkerStats, error) {

	ws := WorkerStats{}

	// Récupérer l'utilisation du CPU (en %)
	cpuPercent, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		return ws, fmt.Errorf("erreur lors de la récupération de l'utilisation CPU : %v", err)
	}
	ws.CPUUsage = cpuPercent[0]

	// Récupérer l'utilisation de la RAM
	virtualMem, err := mem.VirtualMemory()
	if err != nil {
		return ws, fmt.Errorf("erreur lors de la récupération de l'utilisation RAM : %v", err)
	}
	ws.RAMUsage = virtualMem.UsedPercent

	// Renvoie le pourcentage d'utilisation CPU et RAM
	return ws, nil
}
