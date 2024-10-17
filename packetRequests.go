package main

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/vmihailenco/msgpack"
)

func processPacketRequest(client MQTT.Client, packetRequestPayload []byte) {
	// Unmarshal the packet request
	var packetRequest PacketRequest
	err := json.Unmarshal(packetRequestPayload, &packetRequest)
	if err != nil {
		log.Printf("Failed to unmarshal packet request: %v", err)
		return
	}

	// Read the packet data from disk
	packetData, err := readPacketData(packetRequest.VideoID, packetRequest.PacketNumber)
	if err != nil {
		log.Printf("Failed to read packet data: %v", err)
		return
	}

	// Reconstruct the packet
	packet := VideoPacket{
		VideoID:      packetRequest.VideoID,
		PacketNumber: packetRequest.PacketNumber,
		Data:         packetData,
	}

	// Marshal the packet into msgpack
	packetBytes, err := msgpack.Marshal(packet)
	if err != nil {
		log.Printf("Failed to marshal packet: %v", err)
		return
	}

	// Publish the packet data to the topic videoID-stream
	topic := fmt.Sprintf("%s-stream", packetRequest.VideoID)
	if token := client.Publish(topic, 0, false, packetBytes); token.Wait() && token.Error() != nil {
		log.Printf("Failed to publish packet data: %v", token.Error())
	}
	log.Printf("Published packet %d for video %s to topic %s", packetRequest.PacketNumber, packetRequest.VideoID, topic)
}

func readPacketData(videoID string, packetNumber int) ([]byte, error) {
	// Get the video
	video := videoManager.GetVideo(videoID)

	// Lock the video struct during read
	video.mu.Lock()
	defer video.mu.Unlock()

	// Check if the packet offset is known
	offset, exists := video.Packets[packetNumber]
	if !exists {
		log.Printf("Packet number %d not found for video %s", packetNumber, videoID)
		return nil, fmt.Errorf("packet number %d not found", packetNumber)
	}

	// Open the video file
	videoFilePath := filepath.Join(dataDirectory, fmt.Sprintf("%s.dat", videoID))
	file, err := os.Open(videoFilePath)
	if err != nil {
		log.Printf("Failed to open video file %s: %v", videoFilePath, err)
		return nil, err
	}
	defer file.Close()

	// Seek to the packet offset
	_, err = file.Seek(offset, os.SEEK_SET)
	if err != nil {
		log.Printf("Failed to seek to offset %d: %v", offset, err)
		return nil, err
	}

	// Read packet number (should match)
	var pktNum int32
	err = binary.Read(file, binary.BigEndian, &pktNum)
	if err != nil {
		log.Printf("Failed to read packet number: %v", err)
		return nil, err
	}
	if int(pktNum) != packetNumber {
		log.Printf("Packet number mismatch: expected %d, got %d", packetNumber, pktNum)
		return nil, fmt.Errorf("packet number mismatch")
	}

	// Read data length
	var dataLength int32
	err = binary.Read(file, binary.BigEndian, &dataLength)
	if err != nil {
		log.Printf("Failed to read data length: %v", err)
		return nil, err
	}

	// Read packet data
	data := make([]byte, dataLength)
	_, err = file.Read(data)
	if err != nil {
		log.Printf("Failed to read packet data: %v", err)
		return nil, err
	}

	return data, nil
}
