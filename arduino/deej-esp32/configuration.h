#ifndef CONFIGURATION_H
#define CONFIGURATION_H

// WiFi Configuration
#define SSID        "SSID_NAME"   // Change to your WiFi SSID
#define PASSWORD    "SSID_PWD"    // Change to your WiFi password

// Hostname for the ESP32
#define HOSTNAME    "deej-esp32"

// Destination IP address for UDP communication
#define DEST_IP     "192.168.0.1"  // Change to the IP address of your server

// Slider Configuration
const int NUM_SLIDERS = 5;
const int MIN_CHANGE = 35;
const int analogInputs[NUM_SLIDERS] = {9, 10, 11, 12, 13};

// Analog configuration
#define ANALOG_READ_RESOLUTION 10 // Sampling depth can be set to 10, 11 or 12
#define ANALOG_SET_ATTENUATION ADC_11db // Attenuation values can be ADC_0db, ADC_2_5db, ADC_6db, ADC_11db
float offsetValue = 0.0; // Set your specific offset value here
float multiplierValue = 1.0; // Set your specific multiplier value here

// Button Configuration
const int NUM_BUTTONS = 1;
const int buttonInputs[NUM_BUTTONS] = {14};
#define BUTTON_TIMEOUT 50 // in milliseconds

#endif
