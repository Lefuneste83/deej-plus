#define CONFIGURATION_H

// WiFi Configuration
#define SSID        "SSID_NAME"   // Change to your WiFi SSID
#define PASSWORD    "SSID_PWD"    // Change to your WiFi password

// Hostname for the ESP32
#define HOSTNAME    "deej-esp32"

// Destination IP address for UDP communication
#define DEST_IP     "192.168.0.123"  // Change to the IP address of your server

// Slider Configuration
const int NUM_SLIDERS = 5;
const int MIN_CHANGE = 35;
const int analogPins[NUM_SLIDERS] = {9, 10, 11, 12, 13};

// Button Configuration
const int NUM_BUTTONS = 1;
const int buttonInputs[NUM_BUTTONS] = {14};

#endif
