#include <Adafruit_NeoPixel.h>
#include <WiFi.h>
#include <WiFiUdp.h>
#include "configuration.h"
#include <i2c_adc_ads7828.h>

int analogSliderValues[NUM_SLIDERS] = {0};
int *buttonValues = nullptr;
int *sentButtonValues = nullptr;

int readings[NUM_SLIDERS][NUM_SAMPLES] = {0};
int medianSamples[MEDIAN_SAMPLES];
int readIndex[NUM_SLIDERS] = {0};
int total[NUM_SLIDERS] = {0};
int average[NUM_SLIDERS] = {0};
float smoothedValue[NUM_SLIDERS] = {0};

unsigned long *buttonTimerStart = nullptr; // Pointer to store the start time of the button timer for each button
bool *isButtonPressed = nullptr; // Pointer to indicate if each button is pressed

WiFiUDP Udp;
IPAddress destIp;

// Pattern types supported:
enum  pattern { NONE, RAINBOW_CYCLE, THEATER_CHASE, COLOR_WIPE, SCANNER, FADE };
// Patern directions supported:
enum  direction { FORWARD, REVERSE };

// NeoPattern Class - derived from the Adafruit_NeoPixel class
class NeoPatterns : public Adafruit_NeoPixel
{
    public:

    // Member Variables:  
    pattern  ActivePattern;  // which pattern is running
    direction Direction;     // direction to run the pattern
    
    unsigned long Interval;   // milliseconds between updates
    unsigned long lastUpdate; // last update of position
    
    uint32_t Color1, Color2;  // What colors are in use
    uint16_t TotalSteps;  // total number of steps in the pattern
    uint16_t Index;  // current step within the pattern
    
    void (*OnComplete)();  // Callback on completion of pattern
    
    // Constructor - calls base-class constructor to initialize strip
    NeoPatterns(uint16_t pixels, uint8_t pin, uint8_t type, void (*callback)())
    :Adafruit_NeoPixel(pixels, pin, type)
    {
        OnComplete = callback;
    }
    
    // Update the pattern
    void Update()
    {
        if((millis() - lastUpdate) > Interval) // time to update
        {
            lastUpdate = millis();
            switch(ActivePattern)
            {
                case RAINBOW_CYCLE:
                    RainbowCycleUpdate();
                    break;
                case THEATER_CHASE:
                    TheaterChaseUpdate();
                    break;
                case COLOR_WIPE:
                    ColorWipeUpdate();
                    break;
                case SCANNER:
                    ScannerUpdate();
                    break;
                case FADE:
                    FadeUpdate();
                    break;
                default:
                    break;
            }
        }
    }
  
    // Increment the Index and reset at the end
    void Increment()
    {
        if (Direction == FORWARD)
        {
           Index++;
           if (Index >= TotalSteps)
            {
                Index = 0;
                if (OnComplete != NULL)
                {
                    OnComplete(); // call the comlpetion callback
                }
            }
        }
        else // Direction == REVERSE
        {
            --Index;
            if (Index <= 0)
            {
                Index = TotalSteps-1;
                if (OnComplete != NULL)
                {
                    OnComplete(); // call the comlpetion callback
                }
            }
        }
    }
    
    // Reverse pattern direction
    void Reverse()
    {
        if (Direction == FORWARD)
        {
            Direction = REVERSE;
            Index = TotalSteps-1;
        }
        else
        {
            Direction = FORWARD;
            Index = 0;
        }
    }
    
    // Initialize for a RainbowCycle
    void RainbowCycle(uint8_t interval, direction dir = FORWARD)
    {
        ActivePattern = RAINBOW_CYCLE;
        Interval = interval;
        TotalSteps = 255;
        Index = 0;
        Direction = dir;
    }
    
    // Update the Rainbow Cycle Pattern
    void RainbowCycleUpdate()
    {
        for(int i=0; i< numPixels(); i++)
        {
            setPixelColor(i, Wheel(((i * 256 / numPixels()) + Index) & 255));
        }
        show();
        Increment();
    }

    // Initialize for a Theater Chase
    void TheaterChase(uint32_t color1, uint32_t color2, uint8_t interval, direction dir = FORWARD)
    {
        ActivePattern = THEATER_CHASE;
        Interval = interval;
        TotalSteps = numPixels();
        Color1 = color1;
        Color2 = color2;
        Index = 0;
        Direction = dir;
   }
    
    // Update the Theater Chase Pattern
    void TheaterChaseUpdate()
    {
        for(int i=0; i< numPixels(); i++)
        {
            if ((i + Index) % 3 == 0)
            {
                setPixelColor(i, Color1);
            }
            else
            {
                setPixelColor(i, Color2);
            }
        }
        show();
        Increment();
    }

    // Initialize for a ColorWipe
    void ColorWipe(uint32_t color, uint8_t interval, direction dir = FORWARD)
    {
        ActivePattern = COLOR_WIPE;
        Interval = interval;
        TotalSteps = numPixels();
        Color1 = color;
        Index = 0;
        Direction = dir;
    }
    
    // Update the Color Wipe Pattern
    void ColorWipeUpdate()
    {
        setPixelColor(Index, Color1);
        show();
        Increment();
    }
    
    // Initialize for a SCANNNER
    void Scanner(uint32_t color1, uint8_t interval)
    {
        ActivePattern = SCANNER;
        Interval = interval;
        TotalSteps = (numPixels() - 1) * 2;
        Color1 = color1;
        Index = 0;
    }

    // Update the Scanner Pattern
    void ScannerUpdate()
    { 
        for (int i = 0; i < numPixels(); i++)
        {
            if (i == Index)  // Scan Pixel to the right
            {
                 setPixelColor(i, Color1);
            }
            else if (i == TotalSteps - Index) // Scan Pixel to the left
            {
                 setPixelColor(i, Color1);
            }
            else // Fading tail
            {
                 setPixelColor(i, DimColor(getPixelColor(i)));
            }
        }
        show();
        Increment();
    }
    
    // Initialize for a Fade
    void Fade(uint32_t color1, uint32_t color2, uint16_t steps, uint8_t interval, direction dir = FORWARD)
    {
        ActivePattern = FADE;
        Interval = interval;
        TotalSteps = steps;
        Color1 = color1;
        Color2 = color2;
        Index = 0;
        Direction = dir;
    }
    
    // Update the Fade Pattern
    void FadeUpdate()
    {
        // Calculate linear interpolation between Color1 and Color2
        // Optimise order of operations to minimize truncation error
        uint8_t red = ((Red(Color1) * (TotalSteps - Index)) + (Red(Color2) * Index)) / TotalSteps;
        uint8_t green = ((Green(Color1) * (TotalSteps - Index)) + (Green(Color2) * Index)) / TotalSteps;
        uint8_t blue = ((Blue(Color1) * (TotalSteps - Index)) + (Blue(Color2) * Index)) / TotalSteps;
        
        ColorSet(Color(red, green, blue));
        show();
        Increment();
    }
   
    // Calculate 50% dimmed version of a color (used by ScannerUpdate)
    uint32_t DimColor(uint32_t color)
    {
        // Shift R, G and B components one bit to the right
        uint32_t dimColor = Color(Red(color) >> 1, Green(color) >> 1, Blue(color) >> 1);
        return dimColor;
    }

    // Set all pixels to a color (synchronously)
    void ColorSet(uint32_t color)
    {
        for (int i = 0; i < numPixels(); i++)
        {
            setPixelColor(i, color);
        }
        show();
    }

    // Returns the Red component of a 32-bit color
    uint8_t Red(uint32_t color)
    {
        return (color >> 16) & 0xFF;
    }

    // Returns the Green component of a 32-bit color
    uint8_t Green(uint32_t color)
    {
        return (color >> 8) & 0xFF;
    }

    // Returns the Blue component of a 32-bit color
    uint8_t Blue(uint32_t color)
    {
        return color & 0xFF;
    }
    
    // Input a value 0 to 255 to get a color value.
    // The colours are a transition r - g - b - back to r.
    uint32_t Wheel(byte WheelPos)
    {
        WheelPos = 255 - WheelPos;
        if(WheelPos < 85)
        {
            return Color(255 - WheelPos * 3, 0, WheelPos * 3);
        }
        else if(WheelPos < 170)
        {
            WheelPos -= 85;
            return Color(0, WheelPos * 3, 255 - WheelPos * 3);
        }
        else
        {
            WheelPos -= 170;
            return Color(WheelPos * 3, 255 - WheelPos * 3, 0);
        }
    }
};

void Ring1Complete();

// Define some NeoPatterns for the two rings and the stick
//  as well as some completion routines
NeoPatterns Ring1(LED_NUMBER, LED_PIN, NEO_GRB + NEO_KHZ800, &Ring1Complete);

// Initialize everything and prepare to start
void setup() {
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
   
  // Initialize all the pixelStrips
  Ring1.begin();
   
  // Kick off a pattern
//  Ring1.TheaterChase(Ring1.Color(255,255,0), Ring1.Color(0,0,50), 100);
  neopixelWrite(RGB_BUILTIN, 0, 0, 0);  // Off / black
}

// Main loop
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

  if (sliderChange) {
      sendSliderValues(analogSliderValues);
      neopixelWrite(RGB_BUILTIN, 0, 0, 0);  // Off / black
      delay(25);
  }
  if (buttonChange) {
      sendButtonValues(sentButtonValues);
  }

  //sendValues("Slider", analogSliderValues, updateSliderValues());
  //sendValues("Button", sentButtonValues, checkButtonStates());

  // Update the rings.
  Ring1.Update();
  Ring1.ActivePattern = LED_EFFECT;
  Ring1.Interval = interval;
  
}

bool updateSliderValues() {
  bool valueChange = false;
  for (int potIndex = 0; potIndex < NUM_SLIDERS; potIndex++) {
    // Take a new reading
    int newValue = analogRead(analogInputs[potIndex]);

    // Update the total for averaging
    total[potIndex] -= readings[potIndex][readIndex[potIndex]];
    readings[potIndex][readIndex[potIndex]] = newValue;
    total[potIndex] += readings[potIndex][readIndex[potIndex]];
    readIndex[potIndex] = (readIndex[potIndex] + 1) % NUM_SAMPLES;

    // Calculate the average
    average[potIndex] = total[potIndex] / NUM_SAMPLES;

    // Apply the low-pass filter
    smoothedValue[potIndex] = alpha * average[potIndex] + (1 - alpha) * smoothedValue[potIndex];

    // Collect samples for median filtering
    for (int i = 0; i < MEDIAN_SAMPLES; i++) {
      medianSamples[i] = analogRead(analogInputs[potIndex]);
      delay(2); // Small delay between samples
    }

    // Calculate the median
    int medianValue = median(medianSamples, MEDIAN_SAMPLES);

    // Combine the results
    int combinedValue = (smoothedValue[potIndex] + medianValue) / 2;

    // Apply the offset and multiplier
    combinedValue = (combinedValue * multiplierValue) + offsetValue;

    // Ensure the combined value is within valid range
    if (combinedValue > analogSliderValues[potIndex] + MIN_CHANGE || combinedValue < analogSliderValues[potIndex] - MIN_CHANGE) {
      if(combinedValue+MIN_CHANGE>1023){
        analogSliderValues[potIndex] = 1023;}

      else if(combinedValue-MIN_CHANGE<0){
        analogSliderValues[potIndex] = 0;}

      else{
        analogSliderValues[potIndex] = combinedValue;}

      valueChange = true;
    }
  }
  return valueChange;
}

//    for (int i = 0; i < NUM_SLIDERS; i++) {
//        int sum = 0;
//        for (byte n = 0; n < NUM_SAMPLES; n++) {
//            sum += analogRead(analogInputs[i]);
//            delay(2);
//        }
//        int rawVal = sum / NUM_SAMPLES;

        // Apply offset and multiplier
//        float correctedVal = (rawVal + offsetValue) * multiplierValue;

//        // Ensure corrected value stays within valid range
//        int newVal = constrain(correctedVal, 0, 1023);
//        int oldVal = analogSliderValues[i];

//        if (abs(newVal - oldVal) >= MIN_CHANGE) {
//            analogSliderValues[i] = newVal;
//            valueChange = true;
//        }
//    }
//    return valueChange;
//}

int median(int *array, int size) {
  int temp;
  for (int i = 0; i < size - 1; i++) {
    for (int j = i + 1; j < size; j++) {
      if (array[j] < array[i]) {
        temp = array[i];
        array[i] = array[j];
        array[j] = temp;
      }
    }
  }
  return array[size / 2];
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
                neopixelWrite(RGB_BUILTIN, 0, 0, 0);  // Off / black
                delay(100);
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

void sendSliderValues(int* sliderValues) {
    String builtString = "Slider";
    for (int i = 0; i < NUM_SLIDERS; i++) {
        builtString += "|" + String(sliderValues[i]);
    }
    builtString += "\r\n";

    Serial.println(builtString);

    Udp.beginPacket(destIp, 16990);
    Udp.write((uint8_t *)builtString.c_str(), builtString.length());
    Udp.endPacket();
}

void sendButtonValues(int* buttonValues) {
    String builtString = "Button";
    for (int i = 0; i < NUM_BUTTONS; i++) {
        builtString += "|" + String(buttonValues[i]);
    }
    builtString += "\r\n";

    Serial.println(builtString);

    Udp.beginPacket(destIp, 16990);
    Udp.write((uint8_t *)builtString.c_str(), builtString.length());
    Udp.endPacket();
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

//------------------------------------------------------------
//Completion Routines - get called on completion of a pattern
//------------------------------------------------------------

// Ring1 Completion Callback
void Ring1Complete()
{
    if (digitalRead(14) == LOW)  // Button #2 pressed
    {
        // Alternate color-wipe patterns with Ring2
        Ring1.Color1 = Ring1.Wheel(random(255));
        Ring1.Interval = 20000;
    }
    else  // Retrn to normal
    {
      Ring1.Reverse();
    }
}
