/*
IoT Log Simulator for Edge Insights

PURPOSE:
This simulator generates realistic IoT device logs and sends them to the Edge Insights
WebSocket server for storage in TimescaleDB. It's designed to test the end-to-end
log ingestion pipeline and demonstrate real-time IoT data processing.

FEATURES:
- Simulates multiple IoT device types (sensors, cameras, controllers)
- Generates realistic log messages with different severity levels
- Sends logs via WebSocket to the Edge Insights server
- Configurable simulation parameters (device count, frequency, duration)
- Real-time feedback on log transmission success/failure

DEVICE TYPES SIMULATED:
- Temperature sensors (with realistic temperature readings)
- Humidity sensors (with realistic humidity percentages)
- Motion detectors (with motion detection events)
- Cameras (with recording and storage events)
- Controllers (with system status updates)

LOG TYPES GENERATED:
- INFO: Normal operational messages
- WARN: Warning conditions
- ERROR: Error conditions and malfunctions
- DEBUG: Debugging information

USAGE:
Run this simulator while the Edge Insights server is running to populate
the TimescaleDB database with realistic IoT log data for testing and
demonstration purposes.

AUTHOR: Edge Insights Project
VERSION: 1.0
*/

package main

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"time"

	"edge-insights/internal/ws"

	"github.com/gorilla/websocket"
)

// SimulatorConfig holds configuration for the log simulator
// This struct defines how the simulation will behave
type SimulatorConfig struct {
	WebSocketURL string        // URL of the WebSocket server to connect to
	DeviceCount  int           // Number of IoT devices to simulate
	LogInterval  time.Duration // How often to send logs (e.g., every 2 seconds)
	Duration     time.Duration // How long to run the simulation
}

// Device represents a simulated IoT device
// Each device has an ID, type, and location for realistic log generation
type Device struct {
	ID       string // Unique device identifier (e.g., "device_001")
	Type     string // Type of IoT device (e.g., "temperature_sensor")
	Location string // Physical location of the device (e.g., "warehouse_a")
}

// LogSimulator manages the simulation of IoT devices
// This is the main orchestrator that coordinates device simulation and log sending
type LogSimulator struct {
	config  SimulatorConfig // Configuration for the simulation
	devices []Device        // List of simulated devices
	conn    *websocket.Conn // WebSocket connection to the server
}

// main is the entry point of the log simulator
// It sets up configuration, creates the simulator, and runs the simulation
func main() {
	// Define simulation parameters
	config := SimulatorConfig{
		WebSocketURL: "ws://localhost:8080/ws", // Connect to our WebSocket server
		DeviceCount:  5,                        // Simulate 5 different IoT devices
		LogInterval:  2 * time.Second,          // Send a log every 2 seconds
		Duration:     30 * time.Second,         // Run simulation for 30 seconds
	}

	// Create a new simulator instance with our configuration
	simulator := NewLogSimulator(config)

	// Log startup information
	log.Println("Starting IoT Log Simulator...")
	log.Printf("Connecting to: %s", config.WebSocketURL)
	log.Printf("Simulating %d devices", config.DeviceCount)
	log.Printf("Log interval: %v", config.LogInterval)
	log.Printf("Duration: %v", config.Duration)

	// Run the simulation and handle any errors
	if err := simulator.Run(); err != nil {
		log.Fatalf("Simulator failed: %v", err)
	}

	log.Println("Simulation completed successfully!")
}

// NewLogSimulator creates a new log simulator with the given configuration
// It initializes the simulator and generates the list of devices to simulate
func NewLogSimulator(config SimulatorConfig) *LogSimulator {
	return &LogSimulator{
		config:  config,
		devices: generateDevices(config.DeviceCount), // Create the device list
	}
}

// generateDevices creates realistic IoT devices for simulation
// This function creates a variety of device types and locations to make logs realistic
func generateDevices(count int) []Device {
	// Define different types of IoT devices we want to simulate
	deviceTypes := []string{"temperature_sensor", "humidity_sensor", "motion_detector", "camera", "controller"}

	// Define different locations where devices might be deployed
	locations := []string{"warehouse_a", "warehouse_b", "office_floor_1", "parking_lot", "server_room"}

	// Create the device list
	devices := make([]Device, count)
	for i := 0; i < count; i++ {
		devices[i] = Device{
			ID:       fmt.Sprintf("device_%03d", i+1), // Generate IDs like "device_001", "device_002"
			Type:     deviceTypes[i%len(deviceTypes)], // Cycle through device types
			Location: locations[i%len(locations)],     // Cycle through locations
		}
	}
	return devices
}

// Run starts the simulation and manages the main simulation loop
// This is the core function that coordinates log generation and sending
func (s *LogSimulator) Run() error {
	// Connect to the WebSocket server
	// This establishes the connection that will be used to send all logs
	conn, _, err := websocket.DefaultDialer.Dial(s.config.WebSocketURL, nil)
	if err != nil {
		return fmt.Errorf("failed to connect to WebSocket: %w", err)
	}
	defer conn.Close() // Ensure connection is closed when function exits
	s.conn = conn

	log.Println("Connected to WebSocket server")

	// Create a timer to stop the simulation after the specified duration
	timer := time.NewTimer(s.config.Duration)
	defer timer.Stop()

	// Create a ticker to trigger log sending at regular intervals
	ticker := time.NewTicker(s.config.LogInterval)
	defer ticker.Stop()

	// Main simulation loop
	logCount := 0
	for {
		select {
		case <-timer.C:
			// Simulation time is up, stop and report results
			log.Printf("Simulation completed. Sent %d logs", logCount)
			return nil
		case <-ticker.C:
			// Time to send another log
			if err := s.sendRandomLog(); err != nil {
				log.Printf("Error sending log: %v", err)
			} else {
				logCount++ // Count successful log sends
			}
		}
	}
}

// sendRandomLog generates and sends a random log message
// This function creates realistic log data and sends it to the server
func (s *LogSimulator) sendRandomLog() error {
	// Pick a random device from our device list
	device := s.devices[rand.Intn(len(s.devices))]

	// Generate a realistic log message for this device
	logMessage := s.generateLogMessage(device)

	// Convert the log message to JSON format for transmission
	jsonData, err := json.Marshal(logMessage)
	if err != nil {
		return fmt.Errorf("failed to marshal log: %w", err)
	}

	// Send the JSON log message via WebSocket
	if err := s.conn.WriteMessage(websocket.TextMessage, jsonData); err != nil {
		return fmt.Errorf("failed to send log: %w", err)
	}

	// Wait for and read the server's response
	var response ws.LogResponse // Uses the LogResponse struct from types.go
	if err := s.conn.ReadJSON(&response); err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	// Log the result of the operation
	if response.Success {
		log.Printf("✓ Log sent: %s - %s", device.ID, logMessage.Message)
	} else {
		log.Printf("✗ Log failed: %s", response.Error)
	}

	return nil
}

// generateLogMessage creates realistic log messages based on device type
// This function generates different types of logs (INFO, WARN, ERROR, DEBUG)
// with realistic content based on the device type
func (s *LogSimulator) generateLogMessage(device Device) ws.LogMessage {
	// Define possible log types and their probabilities
	logTypes := []string{"INFO", "WARN", "ERROR", "DEBUG"}
	logType := logTypes[rand.Intn(len(logTypes))]

	var message string

	// Generate device-specific messages based on device type
	switch device.Type {
	case "temperature_sensor":
		temp := 15 + rand.Float64()*25 // Generate temperature between 15-40°C
		if logType == "ERROR" {
			message = fmt.Sprintf("Temperature sensor malfunction at %s", device.Location)
		} else {
			message = fmt.Sprintf("Temperature reading: %.1f°C at %s", temp, device.Location)
		}
	case "humidity_sensor":
		humidity := 30 + rand.Float64()*50 // Generate humidity between 30-80%
		if logType == "ERROR" {
			message = fmt.Sprintf("Humidity sensor calibration error at %s", device.Location)
		} else {
			message = fmt.Sprintf("Humidity reading: %.1f%% at %s", humidity, device.Location)
		}
	case "motion_detector":
		if logType == "INFO" {
			message = fmt.Sprintf("Motion detected at %s", device.Location)
		} else if logType == "ERROR" {
			message = fmt.Sprintf("Motion sensor offline at %s", device.Location)
		} else {
			message = fmt.Sprintf("Motion sensor status check at %s", device.Location)
		}
	case "camera":
		if logType == "INFO" {
			message = fmt.Sprintf("Camera recording started at %s", device.Location)
		} else if logType == "ERROR" {
			message = fmt.Sprintf("Camera storage full at %s", device.Location)
		} else {
			message = fmt.Sprintf("Camera maintenance check at %s", device.Location)
		}
	case "controller":
		if logType == "INFO" {
			message = fmt.Sprintf("System check completed at %s", device.Location)
		} else if logType == "ERROR" {
			message = fmt.Sprintf("Controller communication timeout at %s", device.Location)
		} else {
			message = fmt.Sprintf("Controller status update at %s", device.Location)
		}
	default:
		message = fmt.Sprintf("Device activity at %s", device.Location)
	}

	// Return the complete log message using the LogMessage struct from types.go
	return ws.LogMessage{
		Time:     time.Now(), // Current timestamp
		DeviceID: device.ID,  // Device identifier
		LogType:  logType,    // Type of log (INFO, WARN, ERROR, DEBUG)
		Message:  message,    // The actual log message
	}
}
