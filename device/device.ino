// include the WiFi library and HTTPClient
#include <WiFi.h>
#include <HTTPClient.h>
// include json library (https://github.com/bblanchon/ArduinoJson)
#include <ArduinoJson.h>
// include MQTT library
#include <PubSubClient.h>
// eeprom lib has been deprecated for esp32, the recommended way is to use Preferences
#include <Preferences.h>
// IRremoteESP8266 library (https://github.com/crankyoldgit/IRremoteESP8266)
#include "PinDefinitionsAndMore.h"
#include <IRremote.h>
// #include <IRremoteESP8266.h>
// #include <IRsend.h>

// config IRremoteESP8266
// const uint16_t irGpio = 4;  // ESP8266 GPIO pin to use. Recommended: 4 (D2).
// IRsend irsend(irGpio);  // Set the GPIO to be used to sending the message.
// const uint16_t NEC_KHZ = 38;

#include "secrets.h"

void callbackMqtt(char* topic, byte* payload, unsigned int length);

// ------------------------------------------------------
// ----------------------- WIFI -------------------------
const char* ssid = SECRET_SSID; 
const char* password = SECRET_PASS;
const char* serverName = "http://192.168.1.71:8082/api/register";
WiFiClient client;

// -----------------------------------------------------
// ---------------------- MQTT -------------------------
const int serverPortMqtt = 1883;
IPAddress serverMqtt(192, 168, 1, 71);
PubSubClient mqttClient(serverMqtt, serverPortMqtt, callbackMqtt, client);

String savedUuid;
Preferences preferences;

void subscribeDevices(const char* command) {
  const char* devices = "devices/";
  const uint totalLength = sizeof(char) * (strlen(devices) + savedUuid.length() + strlen(command) + 1);
  char* topic = (char*)malloc(totalLength);
  strlcpy(topic, devices, totalLength);
  strlcat(topic, savedUuid.c_str(), totalLength);
  strlcat(topic, command, totalLength);
  Serial.print("subscribeDevices - topic: ");
  Serial.println(topic);
  mqttClient.subscribe(topic);
}

void reconnect() { 
  while (!mqttClient.connected()) {
    Serial.println("reconnect - Attempting MQTT connection...");
    mqttClient.setBufferSize(4096);
    
    if (mqttClient.connect("arduinoClient")) {
      Serial.print("reconnect - Connected and subscribing with savedUuid: ");
      Serial.println(savedUuid);
      // subscribe
      subscribeDevices("/onoff");
      subscribeDevices("/temperature");
      subscribeDevices("/mode");
      subscribeDevices("/fanMode");
      subscribeDevices("/fanSpeed");
    } else {
      Serial.print("reconnect - failed, rc=");
      Serial.print(mqttClient.state());
      Serial.println(" - try again in 5 seconds");
      // Wait 5 seconds before retrying
      delay(5000);
    }
  }
}

// void notifyValue(char* type, char* uuid) {
//   char payloadToSend[20];
//   DynamicJsonDocument payloadMsg(256);
//   switch(type) {
//     case "onoff":
//       payloadMsg["on"] = true;
//       break;
//     case "temperature":
//       break;
//     case "mode":
//       break;
//     case "fanspeed":
//       break;
//     case "fanmode":
//       break;
//   }
//   serializeJson(payloadMsg, payloadToSend);

//   mqttClient.publish("devices/4d28731c-fd5a-420e-9e46-8816de6d053d/notify/onoff", payloadToSend);
// }

void callbackMqtt(char* topic, byte* payload, unsigned int length) {
  Serial.println("callbackMqtt - called");
  const uint16_t irOn[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550 };
  const uint16_t irOff[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550 };
  const uint16_t irTemperature24[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550 };
  const uint16_t irTemperature25[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,550,550,550,550,1650,550,550,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,550,550,550,550,1650,550,550,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550 };
  const uint16_t irTemperature26[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550 };
  const uint16_t irTemperature27[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550 };
  const uint16_t irTemperature28[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550 };
  const uint16_t irTemperature29[] = { };
  const uint16_t irModeAuto[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550 };
  const uint16_t irModeCold[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,5304,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550 };
  const uint16_t irModeDry[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550 };
  const uint16_t irModeHot[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,550,550,1650,550,1650,550 };
  const uint16_t irModeFan[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550 };
  const uint16_t irFanSpeedOff[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550 };
  const uint16_t irFanSpeedLow[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550 };
  const uint16_t irFanSpeedMid[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550 };
  const uint16_t irFanSpeedMax[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550 };

  StaticJsonDocument<250> doc;
  DeserializationError error = deserializeJson(doc, payload);
  if (error) {
    Serial.print("callbackMqtt - deserializeJson() failed: ");
    Serial.println(error.f_str());
    return;
  }
  const char* uuid = doc["uuid"];
  const char* profileToken = doc["profileToken"];
  Serial.print("callbackMqtt - uuid: ");
  Serial.println(uuid);
  Serial.print("callbackMqtt - profileToken: ");
  Serial.println(profileToken);

  // ArduinoJson developer says: "don't use doc.containsKey("on"), instead use doc["on"] != nullptr"
  if(doc["on"] != nullptr) {
    const bool on = doc["on"];
    Serial.print("callbackMqtt - on: ");
    Serial.println(on);
    if (on == 1) {
      Serial.println("callbackMqtt - Sending irOn");
      Serial.flush();
      IrSender.sendRaw(irOn, sizeof(irOn) / sizeof(irOn[0]), NEC_KHZ);
    } else if (on == 0) {
      Serial.println("callbackMqtt - Sending irOff");
      Serial.flush();
      IrSender.sendRaw(irOff, sizeof(irOff) / sizeof(irOff[0]), NEC_KHZ);  
    }
  }
  if(doc["temperature"] != nullptr) {
    const int temperature = doc["temperature"];
    Serial.print("callbackMqtt - temperature: ");
    Serial.println(temperature);
    switch(temperature) {
      case 24:
        Serial.println("callbackMqtt - Sending irTemperature24");
        IrSender.sendRaw(irTemperature24, sizeof(irTemperature24) / sizeof(irTemperature24[0]), NEC_KHZ);
        break;
      case 25:
        Serial.println("callbackMqtt - Sending irTemperature25");
        IrSender.sendRaw(irTemperature25, sizeof(irTemperature25) / sizeof(irTemperature25[0]), NEC_KHZ);
        break;
      case 26:
        Serial.println("callbackMqtt - Sending irTemperature26");
        IrSender.sendRaw(irTemperature26, sizeof(irTemperature26) / sizeof(irTemperature26[0]), NEC_KHZ);
        break;
      case 27:
        Serial.println("callbackMqtt - Sending irTemperature27");
        IrSender.sendRaw(irTemperature27, sizeof(irTemperature27) / sizeof(irTemperature27[0]), NEC_KHZ);
        break;
      case 28:
        Serial.println("callbackMqtt - Sending irTemperature28");
        IrSender.sendRaw(irTemperature28, sizeof(irTemperature28) / sizeof(irTemperature28[0]), NEC_KHZ);
        break;
      default:
        Serial.println("callbackMqtt - Cannot send irTemperature. Unsupported temperature value!");
        break;
    }
  }
  if(doc["mode"] != nullptr) {
    const int mode = doc["mode"];
    Serial.print("callbackMqtt - mode: ");
    Serial.println(mode);
    switch(mode) {
      case 0:
        Serial.println("callbackMqtt - Sending irModeAuto");
        IrSender.sendRaw(irModeAuto, sizeof(irModeAuto) / sizeof(irModeAuto[0]), NEC_KHZ);
        break;
      case 1:
        Serial.println("callbackMqtt - Sending irModeCold");
        IrSender.sendRaw(irModeCold, sizeof(irModeCold) / sizeof(irModeCold[0]), NEC_KHZ);
        break;
      case 2:
        Serial.println("callbackMqtt - Sending irModeDry");
        IrSender.sendRaw(irModeDry, sizeof(irModeDry) / sizeof(irModeDry[0]), NEC_KHZ);
        break;
      case 3:
        Serial.println("callbackMqtt - Sending irModeHot");
        IrSender.sendRaw(irModeHot, sizeof(irModeHot) / sizeof(irModeHot[0]), NEC_KHZ);
        break;
      case 4:
        Serial.println("callbackMqtt - Sending irModeFan");
        IrSender.sendRaw(irModeFan, sizeof(irModeFan) / sizeof(irModeFan[0]), NEC_KHZ);
        break;
      default:
        Serial.println("callbackMqtt - Cannot send irMode. Unsupported mode value!");
        break;
    }
  }
  if(doc["fanSpeed"] != nullptr) {
    const int fanSpeed = doc["fanSpeed"];
    Serial.print("callbackMqtt - fanSpeed: ");
    Serial.println(fanSpeed);
    switch(fanSpeed) {
      case 0:
        Serial.println("callbackMqtt - Sending irFanSpeedOff");
        IrSender.sendRaw(irFanSpeedOff, sizeof(irFanSpeedOff) / sizeof(irFanSpeedOff[0]), NEC_KHZ);
        break;
      case 1:
        Serial.println("callbackMqtt - Sending irFanSpeedLow");
        IrSender.sendRaw(irFanSpeedLow, sizeof(irFanSpeedLow) / sizeof(irFanSpeedLow[0]), NEC_KHZ);
        break;
      case 2:
        Serial.println("callbackMqtt - Sending irFanSpeedMid");
        IrSender.sendRaw(irFanSpeedMid, sizeof(irFanSpeedMid) / sizeof(irFanSpeedMid[0]), NEC_KHZ);
        break;
      case 3:
        Serial.println("callbackMqtt - Sending irFanSpeedMax");
        IrSender.sendRaw(irFanSpeedMax, sizeof(irFanSpeedMax) / sizeof(irFanSpeedMax[0]), NEC_KHZ);
        break;
      default:
        Serial.println("callbackMqtt - Cannot send irFanSpeed. Unsupported fanSpeed value!");
        break;
    }
  }
  if(doc["fanMode"] != nullptr) {
    const int fanMode = doc["fanMode"];
    Serial.print("callbackMqtt - fanMode: ");
    Serial.println(fanMode);
    // TODO TODO TODO implement this
  }
  Serial.println("--------------------------");
}

void registerServer() {
  HTTPClient http;
  http.begin(client, serverName);
  http.addHeader("Content-Type", "application/json; charset=utf-8");
  String macAddress = WiFi.macAddress();
  String httpRequestData = "{\"mac\": \"" + WiFi.macAddress() + 
    "\",\"name\": \"" + NAME + 
    "\",\"manufacturer\": \"" + MANUFACTURER +
    "\",\"model\": \"" + MODEL +
    "\",\"type\": \"" + TYPE +
    "\",\"APIToken\": \"" + API_TOKEN + "\"}";
  const int httpResponseCode = http.POST(httpRequestData);
  if (httpResponseCode <= 0) {
    Serial.print("registerServer - Error on sending POST with httpResponseCode = ");
    Serial.println(httpResponseCode);
    http.end();

    Serial.println("registerServer - Retrying in 3 seconds...");
    delay(60000);
    registerServer();
    return;
  }

  Serial.print("registerServer - httpResponseCode = ");
  Serial.println(httpResponseCode);

  if (httpResponseCode == HTTP_CODE_OK) {
    Serial.println("registerServer - HTTP_CODE_OK");
    StaticJsonDocument<500> staticDoc;
    DeserializationError err = deserializeJson(staticDoc, http.getStream());
    // There is no need to check for specific reasons,
    // because err evaluates to true/false in this case,
    // as recommended by the developer of ArduinoJson
    if (!err) {
      Serial.println("registerServer - Deserialization succeeded!");
      serializeJsonPretty(staticDoc, Serial);
      const char* uuidValue = staticDoc["uuid"];
      const char* macValue = staticDoc["mac"];
      const char* nameValue = staticDoc["name"];
      const char* manufacturerValue = staticDoc["manufacturer"];
      const char* modelValue = staticDoc["model"];
      Serial.print("registerServer - uuidValue: ");
      Serial.println(uuidValue);
      Serial.print("registerServer - macValue: ");
      Serial.println(macValue);
      Serial.print("registerServer - nameValue: ");
      Serial.println(nameValue);
      Serial.print("registerServer - manufacturerValue: ");
      Serial.println(manufacturerValue);
      Serial.print("registerServer - modelValue: ");
      Serial.println(modelValue);

      // TODO TODO TODO TODO TODO TODO validate uuidValue. It must have a certain format and it must be a string


      // if (strcmp(macAddress.c_str(), macValue) != 0) {
          // strcmp(NAME, nameValue) != 0 ||
          // strcmp(MANUFACTURER, manufacturerValue) != 0 ||
          // strcmp(MODEL, modelValue) != 0) {
      //   Serial.println("--- ERROR : Request and response data don't match ---");
      //   return;
      // }
      // if (!macAddress.equals(macValue) || nameValue != NAME || manufacturerValue != MANUFACTURER || modelValue != MODEL) {
      //   Serial.println("--- ERROR : Request and response data don't match ---");
      //   return;
      // }

      preferences.begin("ac", false); 
      String uuidStr = String(uuidValue);
      size_t len = preferences.putString("uuid", uuidStr);
      preferences.end();
      if (len != strlen(uuidValue)) {
        Serial.println("************* ERROR **************");
        Serial.println("setup - Cannot SAVE UUID in Preferences");
        Serial.println("**********************************");
      }
    } else {
      Serial.println("cannot deserialize register JSON payload");
    }
  }
  if (httpResponseCode == HTTP_CODE_CONFLICT) {
    Serial.println("registerServer - HTTP_CODE_CONFLICT - already registered");
  }
}

void setup() {
  Serial.begin(115200);
  // To be able to connect Serial monitor after reset or power up and before first print out. Do not wait for an attached Serial Monitor!
  delay(4000);

  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }

  Serial.println("");
  Serial.println("setup - WiFi connected!");
  Serial.print("setup - IP address: ");
  Serial.println(WiFi.localIP());
  Serial.print("setup - MAC address: ");
  Serial.println(WiFi.macAddress());

  Serial.println("setup - Registering this device...");
  registerServer();
  Serial.println("setup - Registration succedeed!");

  preferences.begin("ac", false); 
  savedUuid = preferences.getString("uuid", "");
  preferences.end();

  if (savedUuid.equals("")) {
    Serial.println("************* ERROR **************");
    Serial.println("setup - Cannot read saved UUID from Preferences");
    Serial.println("**********************************");
    return;
  }
 
  // irsend.begin();
  
  delay(1500);
}

void loop() {
  if (!mqttClient.connected()) {
    Serial.println("loop - RECONNECTING...");
    reconnect();
  }
  mqttClient.loop();

  delay(1000);
}