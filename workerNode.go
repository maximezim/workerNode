package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"sync"
	"syscall"
	"time"

	MQTT "github.com/eclipse/paho.mqtt.golang"
	"github.com/gorilla/websocket"
)

var (
	masterNodeAddress = "0.0.0.0:8080"
	mqttBrokerURI     = "tcp://remicaulier.fr:1883"
	mqttClientID      = "worker-node"
	mqttUsername      = "viewer"
	mqttPassword      = "zimzimlegoat"
	dataDirectory     = "./data"
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

var videoManager = NewVideoManager()

func main() {
	if _, err := os.Stat(dataDirectory); os.IsNotExist(err) {
		err = os.Mkdir(dataDirectory, os.ModePerm)
		if err != nil {
			log.Fatalf("Failed to create data directory: %v", err)
		}
	}

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	var wg sync.WaitGroup

	mqttClient := connectToMQTTBroker()
	defer mqttClient.Disconnect(250)

	// Handle reconnections to master node
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			wsConn := connectToMasterNode()
			receiveDataFromMaster(wsConn)
			wsConn.Close()
			log.Println("Disconnected from master node, retrying in 5 seconds...")
			time.Sleep(5 * time.Second)
		}
	}()

	// Wait for interrupt signal
	<-sigChan
	log.Println("Interrupt signal received, shutting down...")

	wg.Wait()
	log.Println("Shutdown complete")
}

func connectToMQTTBroker() MQTT.Client {
	opts := MQTT.NewClientOptions()
	opts.AddBroker(mqttBrokerURI)
	opts.SetClientID(mqttClientID + "-" + generateClientID())
	opts.SetUsername(mqttUsername)
	opts.SetPassword(mqttPassword)
	opts.OnConnectionLost = func(client MQTT.Client, err error) {
		log.Printf("MQTT connection lost: %v", err)
	}
	opts.OnReconnecting = func(client MQTT.Client, opts *MQTT.ClientOptions) {
		log.Printf("Reconnecting to MQTT broker...")
	}

	client := MQTT.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatalf("Error connecting to MQTT broker: %v", token.Error())
	}
	log.Println("Connected to MQTT broker")
	return client
}

func connectToMasterNode() *websocket.Conn {
	u := url.URL{Scheme: "ws", Host: masterNodeAddress, Path: "/ws"}
	log.Printf("Connecting to master node at %s", u.String())

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Printf("Failed to connect to master node: %v", err)
		return nil
	}
	log.Println("Connected to master node")

	// Receive assigned name
	workerName, err := receiveAssignedName(conn)
	if err != nil {
		log.Printf("Failed to receive worker name: %v", err)
	} else {
		log.Printf("Assigned worker name: %s", workerName)
	}
	return conn
}

func receiveAssignedName(conn *websocket.Conn) (string, error) {
	_, message, err := conn.ReadMessage()
	if err != nil {
		return "", err
	}
	var data map[string]string
	err = json.Unmarshal(message, &data)
	if err != nil {
		return "", err
	}
	return data["worker_name"], nil
}

func receiveDataFromMaster(conn *websocket.Conn) {
	if conn == nil {
		log.Println("Connection to master node is nil")
		return
	}
	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			log.Printf("Error reading from master node: %v", err)
			break
		}
		var packet VideoPacket
		err = json.Unmarshal(message, &packet)
		if err != nil {
			log.Printf("Error unmarshalling data: %v", err)
			continue
		}
		err = processPacket(packet)
		if err != nil {
			log.Printf("Error processing packet: %v", err)
		}
	}
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

func generateClientID() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
