#include <WiFi.h>
#include <WiFiUdp.h>
#include "configuration.h"

int analogSliderValues[NUM_SLIDERS] = {0};
int buttonValues[NUM_BUTTONS];
int sentButtonValues[NUM_BUTTONS];

WiFiUDP Udp;
IPAddress destIp;

void setup() {
    pinMode(LED_BUILTIN, OUTPUT);

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

    pinMode(14, INPUT_PULLUP);

    Serial.println("Ready");
    Serial.print("IP address: ");
    Serial.println(WiFi.localIP());

    analogReadResolution(10);
    analogSetAttenuation(ADC_0db);
}

void loop() {
    sendValues("Slider", analogSliderValues, updateSliderValues());
    sendValues("Button",sentButtonValues,  checkButtonStates());
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
        int newVal = sum / 5;
        int oldVal = analogSliderValues[i];
        // Checks if the new value is higher or lower than the MIN_CHANGE
        if (abs(newVal - oldVal) >= MIN_CHANGE) {
            if (newVal > 1023 - MIN_CHANGE) {
                analogSliderValues[i] = 1023;
            } else if (newVal < MIN_CHANGE) {
                analogSliderValues[i] = 0;
            } else {
                analogSliderValues[i] = newVal;
            }
            valueChange = true;
        }
    }
    return valueChange;
}

bool checkButtonStates(){
  bool valueChange = false;
  for(int i = 0; i < NUM_BUTTONS; i++){
    int newVal = digitalRead(buttonInputs[i]);
    int oldVal = buttonValues[i];
    if(newVal == 1 && oldVal == 0){
      sentButtonValues[i] = 1;
      valueChange = true;
    }
    else{
      sentButtonValues[i] = 0;
    }
    buttonValues[i] = newVal;
  }
  return valueChange;
}

void sendValues(String type, int* values, bool valueChange) {
    if (valueChange) {
        String builtString = "";

        for (int i = 0; i < NUM_SLIDERS; i++) {
            builtString += String(values[i]);

            if (i < NUM_SLIDERS - 1) {
                builtString += "|";
            }
        }
        Serial.println(builtString + "\n");

        Udp.beginPacket(destIp, 16990);
        Udp.write((uint8_t *)builtString.c_str(), builtString.length());
        Udp.endPacket();
    }
}
