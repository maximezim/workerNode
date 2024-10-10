package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"time"
)

func NewVideoManager() *VideoManager {
	return &VideoManager{
		videos: make(map[string]*Video),
	}
}

func (vm *VideoManager) GetVideo(videoID string) *Video {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	video, exists := vm.videos[videoID]
	if !exists {
		video = &Video{
			VideoID: videoID,
			Packets: make(map[int][]byte),
		}
		vm.videos[videoID] = video
	}
	return video
}

func (vm *VideoManager) RemoveVideo(videoID string) {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	delete(vm.videos, videoID)
}

func assembleAndSaveVideo(video *Video) {
	video.mu.Lock()
	defer video.mu.Unlock()

	// Ensure no double assembly
	if !video.AssemblyComplete {
		return
	}
	video.AssemblyComplete = false // Prevent re-entry

	// Assemble packets in order
	packetNumbers := make([]int, 0, len(video.Packets))
	for num := range video.Packets {
		packetNumbers = append(packetNumbers, num)
	}
	sort.Ints(packetNumbers)
	var videoData []byte
	for _, num := range packetNumbers {
		videoData = append(videoData, video.Packets[num]...)
	}

	// Save videoData to file
	filename := fmt.Sprintf("%s.mp4", video.VideoID)
	filepath := filepath.Join(dataDirectory, filename)
	err := os.WriteFile(filepath, videoData, 0644)
	if err != nil {
		log.Printf("Failed to write video to file: %v", err)
		return
	}
	log.Printf("Video saved to %s", filepath)

	// Remove video from manager
	videoManager.RemoveVideo(video.VideoID)
}

func processPacket(packet VideoPacket) error {
	video := videoManager.GetVideo(packet.VideoID)
	video.mu.Lock()
	defer video.mu.Unlock()

	// Update last packet time for timeout purposes
	video.LastPacketTime = time.Now()

	// Store the packet data if it's new
	if _, exists := video.Packets[packet.PacketNumber]; !exists {
		video.Packets[packet.PacketNumber] = packet.Data
		video.ReceivedPackets++
	}

	// Set the expected total packets if provided
	if packet.TotalPackets > 0 {
		video.ExpectedPackets = packet.TotalPackets
	}

	// Check if all packets have been received
	if (video.ExpectedPackets > 0 && video.ReceivedPackets == video.ExpectedPackets) ||
		(video.ExpectedPackets == 0 && time.Since(video.LastPacketTime) > 5*time.Second) {
		video.AssemblyComplete = true
		go assembleAndSaveVideo(video)
	}

	return nil
}
