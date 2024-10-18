package main

import (
	"time"

	"github.com/shirou/gopsutil/cpu"
)

// getStats renvoie l'utilisation du CPU (%) et de la RAM (%)
func getStats(worker string) (WorkerStats, error) {

	ws := WorkerStats{}
	ws.WorkerName = worker

	// Récupérer l'utilisation du CPU (en %)
	cpuPercent, err := cpu.Percent(1*time.Second, false)
	if err != nil {
		return ws, err
	}
	ws.CPUUsage = cpuPercent[0]

	// 1GB harcodé (osekour)
	ws.MEMUsage = float64(1000000000)

	// Renvoie le pourcentage d'utilisation CPU et RAM
	return ws, nil
}
