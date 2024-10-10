package main

import (
	"sync"
	"time"
)

type VideoPacket struct {
	VideoID      string `json:"video_id"`
	PacketNumber int    `json:"packet_number"`
	TotalPackets int    `json:"total_packets"` // Use 0 if unknown
	Data         []byte `json:"data"`
}

type Video struct {
	VideoID          string
	Packets          map[int][]byte
	ExpectedPackets  int
	ReceivedPackets  int
	LastPacketTime   time.Time
	AssemblyComplete bool
	mu               sync.Mutex
}

type VideoManager struct {
	videos map[string]*Video
	mu     sync.Mutex
}

type WorkerStats struct {
	CPUUsage float64 `json:"cpu_usage"`
	RAMUsage float64 `json:"ram_usage"`
}
