package main

import (
	"encoding/binary"
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
			Packets: make(map[int]int64), // Map packet numbers to file offsets
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
	// Path to the video file
	videoFilePath := filepath.Join(dataDirectory, fmt.Sprintf("%s.dat", packet.VideoID))

	// Open or create the video file
	file, err := os.OpenFile(videoFilePath, os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Printf("Failed to open file for video %s: %v", packet.VideoID, err)
		return err
	}
	defer file.Close()

	// Lock the file to prevent concurrent writes
	video := videoManager.GetVideo(packet.VideoID)
	video.mu.Lock()
	defer video.mu.Unlock()

	// Get the current offset
	offset, err := file.Seek(0, os.SEEK_END)
	if err != nil {
		log.Printf("Failed to seek to end of file: %v", err)
		return err
	}

	// Write packet number and data length
	err = binary.Write(file, binary.BigEndian, int32(packet.PacketNumber))
	if err != nil {
		log.Printf("Failed to write packet number: %v", err)
		return err
	}

	dataLength := int32(len(packet.Data))
	err = binary.Write(file, binary.BigEndian, dataLength)
	if err != nil {
		log.Printf("Failed to write data length: %v", err)
		return err
	}

	// Write packet data
	_, err = file.Write(packet.Data)
	if err != nil {
		log.Printf("Failed to write packet data: %v", err)
		return err
	}

	// Update the packet offset map
	video.Packets[packet.PacketNumber] = offset

	log.Printf("Packet %d for video %s saved at offset %d", packet.PacketNumber, packet.VideoID, offset)
	return nil
}

// HasVideo checks if a video with the given ID exists in the manager.
func (vm *VideoManager) HasVideo(videoID string) bool {
	vm.mu.Lock()
	defer vm.mu.Unlock()
	_, exists := vm.videos[videoID]
	return exists
}
