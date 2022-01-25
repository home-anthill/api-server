// include the WiFi library and HTTPClient
#include <WiFi.h>
#include <HTTPClient.h>
// include json library (https://github.com/bblanchon/ArduinoJson)
#include <ArduinoJson.h>
// include MQTT library
#include <PubSubClient.h>
// eeprom lib has been deprecated for esp32, the recommended way is to use Preferences
#include <Preferences.h>
// IR library
#include "PinDefinitionsAndMore.h"
#include <IRremote.h>

#include "secrets.h"

// -------------------------------------------------------
// -------------------------------------------------------
char ssid[] = SECRET_SSID;        // your network SSID (name)
char password[] = SECRET_PASS;    // your network password (use for WPA, or use as key for WEP)
const char* serverName = "http://192.168.1.71:8082/api/register";
WiFiClient client;

// -------------------------------------------------------
// ---------------------- MQTT -------------------------
const int serverPortMqtt = 1883;
IPAddress serverMqtt(192, 168, 1, 71);

PubSubClient mqttClient(client);

String savedUuid;
Preferences preferences;

void subscribeDevices(const char* command) {
  const char* devices = "devices/";
  const uint devicesLen = strlen(devices);
  const uint savedUuidLen = savedUuid.length();
  const uint commandLen = strlen(command);
  char* topic = (char*)malloc(sizeof(char) * (devicesLen + savedUuidLen + commandLen + 1));
  strcpy(topic, devices);
  strcat(topic, savedUuid.c_str());
  strcat(topic, command);
  Serial.println(topic);
  mqttClient.subscribe(topic);
}

void reconnect() { 
  // Loop until we're reconnected
  while (!mqttClient.connected()) {
    Serial.println("Attempting MQTT connection...");
    // Attempt to connect
    mqttClient.setBufferSize(4096);
    if (mqttClient.connect("arduinoClient")) {
      Serial.print("Connected and subscribing with savedUuid: ");
      Serial.println(savedUuid);
      // subscribe
      subscribeDevices("/onoff");
      subscribeDevices("/temperature");
      subscribeDevices("/mode");
      subscribeDevices("/fanMode");
      subscribeDevices("/fanSpeed");
    } else {
      Serial.print("failed, rc=");
      Serial.print(mqttClient.state());
      Serial.println(" try again in 5 seconds");
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

void callback(char* topic, byte* payload, unsigned int length) {
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

  DynamicJsonDocument doc(300);
  DeserializationError error = deserializeJson(doc, payload);
  if (error) {
    Serial.print(F("deserializeJson() failed: "));
    Serial.println(error.f_str());
    return;
  }
  const char* uuid = doc["uuid"];
  const char* profileToken = doc["profileToken"];
  Serial.println("--------------------------");
  Serial.print("uuid: ");
  Serial.println(uuid);
  Serial.print("profileToken: ");
  Serial.println(profileToken);

  if(doc.containsKey("on")) {
    const bool on = doc["on"];
    Serial.print("on: ");
    Serial.println(on);
    Serial.flush();
    if (on == 1) {
      Serial.println("Sending ON");
      IrSender.sendRaw(irOn, sizeof(irOn) / sizeof(irOn[0]), NEC_KHZ);
    } else if (on == 0) {
      Serial.println("Sending OFF");
      IrSender.sendRaw(irOff, sizeof(irOff) / sizeof(irOff[0]), NEC_KHZ);  
    }
  }
  if(doc.containsKey("temperature")) {
    const int temperature = doc["temperature"];
    Serial.print("temperature: ");
    Serial.println(temperature);
    Serial.flush();
    switch(temperature) {
      case 24:
        IrSender.sendRaw(irTemperature24, sizeof(irTemperature24) / sizeof(irTemperature24[0]), NEC_KHZ);
        break;
      case 25:
        IrSender.sendRaw(irTemperature25, sizeof(irTemperature25) / sizeof(irTemperature25[0]), NEC_KHZ);
        break;
      case 26:
        IrSender.sendRaw(irTemperature26, sizeof(irTemperature26) / sizeof(irTemperature26[0]), NEC_KHZ);
        break;
      case 27:
        IrSender.sendRaw(irTemperature27, sizeof(irTemperature27) / sizeof(irTemperature27[0]), NEC_KHZ);
        break;
      case 28:
        IrSender.sendRaw(irTemperature28, sizeof(irTemperature28) / sizeof(irTemperature28[0]), NEC_KHZ);
        break;
      default:
        Serial.println("Unsupported temperature value");
        break;
    }
  }
  if(doc.containsKey("mode")) {
    const int mode = doc["mode"];
    Serial.println(mode);
    Serial.print("mode: ");
    Serial.println(mode);
    Serial.flush();
    switch(mode) {
      case 0:
        IrSender.sendRaw(irModeAuto, sizeof(irModeAuto) / sizeof(irModeAuto[0]), NEC_KHZ);
        break;
      case 1:
        IrSender.sendRaw(irModeCold, sizeof(irModeCold) / sizeof(irModeCold[0]), NEC_KHZ);
        break;
      case 2:
        IrSender.sendRaw(irModeDry, sizeof(irModeDry) / sizeof(irModeDry[0]), NEC_KHZ);
        break;
      case 3:
        IrSender.sendRaw(irModeHot, sizeof(irModeHot) / sizeof(irModeHot[0]), NEC_KHZ);
        break;
      case 4:
        IrSender.sendRaw(irModeFan, sizeof(irModeFan) / sizeof(irModeFan[0]), NEC_KHZ);
        break;
      default:
        Serial.println("Unsupported mode value");
        break;
    }
  }
  if(doc.containsKey("fanSpeed")) {
    const int fanSpeed = doc["fanSpeed"];
    Serial.println(fanSpeed);
    Serial.print("fanSpeed: ");
    Serial.println(fanSpeed);
    Serial.flush();
    switch(fanSpeed) {
      case 0:
        IrSender.sendRaw(irFanSpeedOff, sizeof(irFanSpeedOff) / sizeof(irFanSpeedOff[0]), NEC_KHZ);
        break;
      case 1:
        IrSender.sendRaw(irFanSpeedLow, sizeof(irFanSpeedLow) / sizeof(irFanSpeedLow[0]), NEC_KHZ);
        break;
      case 2:
        IrSender.sendRaw(irFanSpeedMid, sizeof(irFanSpeedMid) / sizeof(irFanSpeedMid[0]), NEC_KHZ);
        break;
      case 3:
        IrSender.sendRaw(irFanSpeedMax, sizeof(irFanSpeedMax) / sizeof(irFanSpeedMax[0]), NEC_KHZ);
        break;
      default:
        Serial.println("Unsupported fan value");
        break;
    }
  }
  if(doc.containsKey("fanMode")) {
    const int fanMode = doc["fanMode"];
    Serial.println(fanMode);
    Serial.print("fanMode: ");
    Serial.println(fanMode);
    Serial.flush();
  }
  Serial.println("--------------------------");
}

void registerServer() {
  HTTPClient http;
  http.begin(client, serverName);
  http.addHeader("Content-Type", "application/json; charset=utf-8");
  http.addHeader("Authorization", "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJpZCI6NjA1NzIwNywibmFtZSI6IlN0ZWZhbm8gQ2FwcGEiLCJleHAiOjE2Mzc3MDk3MDF9.1Du9-D-zmbyblmrpxM9Lw-MUPJkdE99s7p68yHFHvQo");
  String macAddress = WiFi.macAddress();
  String httpRequestData = "{\"mac\": \"" + WiFi.macAddress() + 
    "\",\"name\": \"" + NAME + 
    "\",\"manufacturer\": \"" + MANUFACTURER +
    "\",\"model\": \"" + MODEL +
    "\",\"type\": \"" + TYPE +
    "\",\"apiToken\": \"" + API_TOKEN + "\"}";
  int httpResponseCode = http.POST(httpRequestData);
  if (httpResponseCode>0) {
    // String response = http.getString();
    Serial.println(httpResponseCode); 
    // Serial.println(response);
    StaticJsonDocument<2048> staticDoc;
    const char* uuidValue;
    const char* macValue;
    const char* nameValue;
    const char* manufacturerValue;
    const char* modelValue;
    DeserializationError err = deserializeJson(staticDoc, http.getStream());
    switch (err.code()) {
      case DeserializationError::Ok:
          Serial.println(F("Deserialization succeeded with uuid"));
          uuidValue = staticDoc["uuid"];
          macValue = staticDoc["mac"];
          nameValue = staticDoc["name"];
          manufacturerValue = staticDoc["manufacturer"];
          modelValue = staticDoc["model"];
          Serial.print("uuidValue: ");
          Serial.println(uuidValue);
          Serial.print("macValue: ");
          Serial.println(macValue);
          Serial.print("nameValue: ");
          Serial.println(nameValue);
          Serial.print("manufacturerValue: ");
          Serial.println(manufacturerValue);
          Serial.print("modelValue: ");
          Serial.println(modelValue);
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
          preferences.putString("uuid", uuidValue);
          preferences.end();

          break;
      case DeserializationError::InvalidInput:
          Serial.print(F("Invalid input!"));
          break;
      case DeserializationError::NoMemory:
          Serial.print(F("Not enough memory"));
          break;
      default:
          Serial.print(F("Deserialization failed"));
          break;
    }
  } else {
    Serial.print("Error on sending POST: ");
    Serial.println(httpResponseCode);
    http.end();

    Serial.println("Retrying in 3 seconds...");
    delay(3000);
    registerServer();
 }
}

void setup() {
  Serial.begin(115200);
  delay(4000); // To be able to connect Serial monitor after reset or power up and before first print out. Do not wait for an attached Serial Monitor!

  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }

  Serial.println("");
  Serial.println("WiFi connected");
  Serial.println("IP address: ");
  Serial.println(WiFi.localIP());
  Serial.println("MAC address: ");
  Serial.println(WiFi.macAddress());

  Serial.println("Registering this device...");
  registerServer();
  Serial.println("Registration success!");

  preferences.begin("ac", false); 
  savedUuid = preferences.getString("uuid", "");
  preferences.end();

  if (savedUuid.equals("")) {
    Serial.println("************* ERROR **************");
    Serial.println("Cannot read UUID from Preferences");
    Serial.println("**********************************");
    return;
  }
 

  // Just to know which program is running on my Arduino
  Serial.println(F("START " __FILE__ " from " __DATE__ "\r\nUsing library version " VERSION_IRREMOTE));
  // arduino mega PWM pins
  // 2 - 13, 44 - 46
  // however, with Adafruit wifi shild, these pins are reserved and unusable:
  // Digital pin 3: IRQ for WiFi
  // Digital pin 4: Card Select for SD card
  // Digital pin 5: WiFi enable
  // Digital pin 10: Chip Select for WiFi
  // Digital pins 11, 12, 13 for SPI communication (both WiFi and SD). Even if optional 6-pin SPI header is used, these pins are unavailable for other use. 

  // IrSender.begin(4, DISABLE_LED_FEEDBACK); // Specify send pin and enable feedback LED at default feedback LED pin
  Serial.print(F("Ready to send IR signals at pin "));
  Serial.println(IR_SEND_PIN);
  
  mqttClient.setServer(serverMqtt, serverPortMqtt);
  mqttClient.setCallback(callback);
  
  // Allow the hardware to sort itself out
  delay(1500);
}

void loop() {
  if (!mqttClient.connected()) {
    Serial.println("RECONNECTING...");
    reconnect();
  }
  mqttClient.loop();

  delay(1000);
}