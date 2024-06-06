#include <WiFi.h>
#include <WiFiUdp.h>
#include "configuration.h"

int analogSliderValues[NUM_SLIDERS] = {0};
int *buttonValues = nullptr;
int *sentButtonValues = nullptr;

unsigned long *buttonTimerStart = nullptr; // Pointer to store the start time of the button timer for each button
bool *isButtonPressed = nullptr; // Pointer to indicate if each button is pressed

WiFiUDP Udp;
IPAddress destIp;

void setup() {
    pinMode(LED_BUILTIN, OUTPUT);

    if (NUM_BUTTONS > 0) {
        buttonValues = new int[NUM_BUTTONS];
        sentButtonValues = new int[NUM_BUTTONS];
        buttonTimerStart = new unsigned long[NUM_BUTTONS];
        isButtonPressed = new bool[NUM_BUTTONS]();

       // Initialize button values to 0
        for (int i = 0; i < NUM_BUTTONS; i++) {
            buttonValues[i] = 0;
            sentButtonValues[i] = 0;
            buttonTimerStart[i] = 0;
            isButtonPressed[i] = false;
        }
    }
    
    destIp.fromString(DEST_IP);

    Serial.begin(115200);
    Serial.println();
    Serial.println("Configuring access point...");

    WiFi.mode(WIFI_STA);
    WiFi.disconnect(true);
    WiFi.config(INADDR_NONE, INADDR_NONE, INADDR_NONE);
    WiFi.setHostname(HOSTNAME);
    WiFi.begin(SSID, PASSWORD);

    while (WiFi.waitForConnectResult() != WL_CONNECTED) {
        Serial.println("Connection Failed! Rebooting...");
        delay(5000);
        ESP.restart();
    }

    // Configure analog inputs
    analogReadResolution(ANALOG_READ_RESOLUTION);
    analogSetAttenuation(ANALOG_SET_ATTENUATION);

    // Configure buttons dynamically
    for (int i = 0; i < NUM_BUTTONS; i++) {
        pinMode(buttonInputs[i], INPUT_PULLUP);
    }

    Serial.println("Ready");
    Serial.print("IP address: ");
    Serial.println(WiFi.localIP());
}

void loop() {
    bool sliderChange = updateSliderValues();
    bool buttonChange = checkButtonStates();

    if (NUM_BUTTONS > 0) {
        // Check if the button timer has expired for each button
        for (int i = 0; i < NUM_BUTTONS; i++) {
            if (isButtonPressed[i] && millis() - buttonTimerStart[i] >= BUTTON_TIMEOUT) {
                // Reset button value to 0
                buttonValues[i] = 0;
                sentButtonValues[i] = 0;
                isButtonPressed[i] = false; // Clear the button pressed flag
                buttonChange = true;
            }
        }
    }

    sendValues("Slider", analogSliderValues, sentButtonValues, sliderChange, buttonChange);
    delay(10);
}

bool updateSliderValues() {
    bool valueChange = false;
    for (int i = 0; i < NUM_SLIDERS; i++) {
        int sum = 0;
        for (byte n = 0; n < 5; n++) {
            sum += analogRead(analogInputs[i]);
            delay(2);
        }
        int rawVal = sum / 5;

        // Apply offset and multiplier
        float correctedVal = (rawVal + offsetValue) * multiplierValue;

        // Ensure corrected value stays within valid range
        int newVal = constrain(correctedVal, 0, 1023);
        int oldVal = analogSliderValues[i];

        if (abs(newVal - oldVal) >= MIN_CHANGE) {
            analogSliderValues[i] = newVal;
            valueChange = true;
        }
    }
    return valueChange;
}

bool checkButtonStates() {
    bool valueChange = false;
    if (NUM_BUTTONS > 0) {
        for (int i = 0; i < NUM_BUTTONS; i++) {
            int newVal = digitalRead(buttonInputs[i]);
            int oldVal = buttonValues[i];
            if (newVal == 0 && oldVal == 1) {
                sentButtonValues[i] = 1;
                valueChange = true;
                isButtonPressed[i] = true;
                buttonTimerStart[i] = millis();
            } else if (newVal == 1 && oldVal == 0) { // Button released
                isButtonPressed[i] = false;
                sentButtonValues[i] = 0;
                valueChange = true;
            }
            buttonValues[i] = newVal;
        }
    }
    return valueChange;
}

void sendValues(String type, int* sliderValues, int* buttonValues, bool sliderChange, bool buttonChange) {
    if (sliderChange || buttonChange) {
        String builtString = "";

        // Iterate over slider values
        for (int i = 0; i < NUM_SLIDERS; i++) {
            builtString += String(sliderValues[i]);
            if (i < NUM_SLIDERS - 1) {
                builtString += "|";
            }
        }

        // Iterate over button values
        if (NUM_BUTTONS > 0) {
            for (int i = 0; i < NUM_BUTTONS; i++) {
                builtString += "|" + String(buttonValues[i]);
                if (i < NUM_BUTTONS - 1) {
                    builtString += "|";
                }
            }
        }

        Serial.println(builtString + "\n");

        Udp.beginPacket(destIp, 16990);
        Udp.write((uint8_t *)builtString.c_str(), builtString.length());
        Udp.endPacket();
    }
}

void cleanup() {
    if (buttonValues != nullptr) {
        delete[] buttonValues;
        buttonValues = nullptr;
    }
    if (sentButtonValues != nullptr) {
        delete[] sentButtonValues;
        sentButtonValues = nullptr;
    }
    if (buttonTimerStart != nullptr) {
        delete[] buttonTimerStart;
        buttonTimerStart = nullptr;
    }
    if (isButtonPressed != nullptr) {
        delete[] isButtonPressed;
        isButtonPressed = nullptr;
    }
}
