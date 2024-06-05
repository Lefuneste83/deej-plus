package deej

import (
	"context"
	"fmt"
	"time"

	"go.bug.st/serial"
)

var prevPort string

func GetPortName() (string, error) {
	comP := ""
	ports, err := serial.GetPortsList()
	if err != nil {
		return "", err
	}

	if len(ports) == 0 {
		return "", fmt.Errorf("no serial ports found")
	} else {
		for _, port := range ports {
			fmt.Println("Testing " + port + " now")
			s, e := sendMessage(port, "connect")
			fmt.Println(e)
			fmt.Println("Recived message: " + s)
			if s == "yes" {
				fmt.Println("Found correct port")
				comP = port
				break
			}
		}
	}
	if comP == "" {
		if prevPort != "" {
			fmt.Println("could not find new serial port using previous one")
			return prevPort, nil
		}
		return comP, fmt.Errorf("could not find the serial port")
	}
	prevPort = comP

	return comP, nil
}

func sendMessage(portName string, message string) (string, error) {
	// Create a context with a timeout of 5 seconds
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Channel to receive the response
	responseChan := make(chan string, 1)
	errorChan := make(chan error, 1)

	// Goroutine to handle opening, writing, and reading from the serial port
	go func() {
		// Open the serial port
		port, err := serial.Open(portName, &serial.Mode{})
		if err != nil {
			errorChan <- fmt.Errorf("failed to open serial port: %v", err)
			return
		}
		defer port.Close()

		// Write message to the serial port
		_, err = port.Write([]byte(message))
		if err != nil {
			errorChan <- fmt.Errorf("failed to write to serial port: %v", err)
			return
		}

		// Read response from the serial port
		response := make([]byte, 100) // Assuming response won't exceed 100 bytes
		n, err := port.Read(response)
		if err != nil {
			errorChan <- fmt.Errorf("failed to read from serial port: %v", err)
			return
		}

		responseChan <- string(response[:n])
	}()

	// Wait for response or timeout
	select {
	case response := <-responseChan:
		return response, nil
	case err := <-errorChan:
		return "", err
	case <-ctx.Done():
		return "", fmt.Errorf("timeout reading from serial port")
	}
}
