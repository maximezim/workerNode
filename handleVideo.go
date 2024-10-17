package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
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

func processPacket(packet VideoPacket) error {
	// Create directory for VideoID if it doesn't exist
	videoDir := filepath.Join(dataDirectory, packet.VideoID)
	if _, err := os.Stat(videoDir); os.IsNotExist(err) {
		err := os.Mkdir(videoDir, os.ModePerm)
		if err != nil {
			log.Printf("Failed to create directory for video %s: %v", packet.VideoID, err)
			return err
		}
	}

	// Create filename for the packet
	filename := fmt.Sprintf("%d.dat", packet.PacketNumber)
	filepath := filepath.Join(videoDir, filename)

	// Save the packet data to file
	err := os.WriteFile(filepath, packet.Data, 0644)
	if err != nil {
		log.Printf("Failed to write packet to file: %v", err)
		return err
	}
	log.Printf("Packet %d for video %s saved to %s", packet.PacketNumber, packet.VideoID, filepath)
	return nil
}
