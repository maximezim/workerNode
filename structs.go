package main

import (
	"sync"
	"time"
)

type VideoPacketSIS struct {
	MsgPackPacket []byte `msgpack:"packet"`
	A             []byte `msgpack:"a"`
	V             []byte `msgpack:"v"`
}

type VideoPacket struct {
	VideoID      string `msgpack:"video_id"`
	PacketNumber int    `msgpack:"packet_number"`
	TotalPackets int    `msgpack:"total_packets"` // Use 0 if unknown
	Data         []byte `msgpack:"data"`
}

type Video struct {
	VideoID          string
	Packets          map[int]int64 // Map packet numbers to file offsets
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
	WorkerName string  `json:"worker"`
	CPUUsage   float64 `json:"cpu_usage"`
	MEMUsage   float64 `json:"mem_usage"`
}

type PacketRequest struct {
	VideoID      string `json:"video_id"`
	PacketNumber int    `json:"packet_number"`
	ChannelUUID  string `json:"channel_uuid"`
}
