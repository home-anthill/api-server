// include the WiFi library and HTTPClient
#include <WiFi.h>
#include <HTTPClient.h>
// include json library (https://github.com/bblanchon/ArduinoJson)
#include <ArduinoJson.h>
// include MQTT library
#include <PubSubClient.h>
// eeprom lib has been deprecated for esp32, the recommended way is to use Preferences
#include <Preferences.h>
// include the TimeAlarms library (https://www.arduino.cc/reference/en/libraries/timealarms/)
#include <Time.h>
#include <TimeAlarms.h>

// include libraries:
// - `Grove - Air quality sensor` by Seeed Studio (latest version on `master branch` or
//    commit hash `58e4c0bb5ce1b0c9b8aa1265e9f726025feb34f0` from `https://github.com/Seeed-Studio/Grove_Air_quality_Sensor`).
//    You cannot use the one published on ArduinoIDE Library Manager, because it's outdated and not compatibile with ESP32 devices.
#include "Air_Quality_Sensor.h"


#include "secrets.h"


// ------------------------------------------------------
// -------------------- AirQuality ----------------------
#define AIR_QUALITY_PIN 4 // Digital pin connected to the airQuality sensor
AirQualitySensor sensor(AIRQUALITYPIN);
// global variable used to store the current value read by the sensor
int currentValue = -1; // initial uninitialized value
// ------------------------------------------------------
// ------------------------------------------------------


// Given below is the CA Certificate "ISRG Root X1" by Let's Encrypt.
// Expiration date June 2035.
//
// This ca_cert is applied to WifiClientSecure,
// so it will be used with both HTTPS and MQTTS connections
//
// YOU CAN GET THIS FILE FROM YOUR OS!
// On Linux, it's in /etc/ssl/certs called either 'ISRG_Root_X1.pem' or 'ca-cert-ISRG_Root_X1.pem'.
//
// https://techtutorialsx.com/2017/11/18/esp32-arduino-https-get-request/
const char* ca_cert = \
"-----BEGIN CERTIFICATE-----\n" \
"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\n" \
"-----END CERTIFICATE-----\n";


// ------------------------------------------------------
// ----------------------- WIFI -------------------------
const char* ssid = SECRET_SSID; 
const char* password = SECRET_PASS;
String macAddress;
# if SSL==true
WiFiClientSecure client;
# else 
WiFiClient client;
# endif


// -----------------------------------------------------
// ---------------------- MQTT -------------------------
// Library doc at https://pubsubclient.knolleary.net/api
const char* mqttUrl = MQTT_URL;
const int mqttPort = MQTT_PORT;
PubSubClient mqttClient(mqttUrl, mqttPort, client);


String savedUuid;
Preferences preferences;

char* getRegisterUrl() {
  Serial.println("getRegisterUrl - creating url based on secrets.h config...");
  # if SSL==true
  const char* httpProtocol = "https://";
  # else 
  const char* httpProtocol = "http://";
  # endif
  const char* serverDomain = SERVER_DOMAIN;
  const char* serverPortPrefix = ":";
  const char* serverPort = SERVER_PORT;
  const char* serverPath = SERVER_PATH;
  const uint totalLength = sizeof(char) * (strlen(httpProtocol) + strlen(serverDomain) + strlen(serverPortPrefix) + strlen(serverPort) + strlen(serverPath) + 1);
  char* registerUrl = (char*)malloc(totalLength);
  strlcpy(registerUrl, httpProtocol, totalLength);
  strlcat(registerUrl, serverDomain, totalLength);
  strlcat(registerUrl, serverPortPrefix, totalLength);
  strlcat(registerUrl, serverPort, totalLength);
  strlcat(registerUrl, serverPath, totalLength);
  return registerUrl;
}

void reconnect() { 
  while (!mqttClient.connected()) {
    Serial.println("reconnect - attempting MQTT connection...");
    mqttClient.setBufferSize(4096);
    
    // here you can use the version with `connect(const char* id, const char* user, const char* pass)` with authentication
    const char* idClient = savedUuid.c_str();
    Serial.print("reconnect - connecting to MQTT with client id = ");
    Serial.println(idClient);

    if (mqttClient.connect(idClient)) {
      Serial.print("reconnect - connected and subscribing with savedUuid: ");
      Serial.println(savedUuid);
    } else {
      Serial.print("reconnect - failed, rc=");
      Serial.print(mqttClient.state());
      Serial.println(" - try again in 5 seconds");
      // Wait 5 seconds before retrying
      delay(5000);
    }
  }
}

void notifyValue(char* type, float value) {
  Serial.println("notifyValue - called with type=" + String(type));
  // check if type is supported
  if (strcmp(type, "airquality") != 0) {
    Serial.println("notifyValue - Cannot send data. Unsupported type=" + String(type));
    return;
  }

  char payloadToSend[562];
  DynamicJsonDocument innerPayloadMsg(50);
  innerPayloadMsg["value"] = value;
  DynamicJsonDocument payloadMsg(512);
  payloadMsg["uuid"] = savedUuid;
  payloadMsg["apiToken"] = API_TOKEN;
  payloadMsg["payload"] = innerPayloadMsg;
  serializeJson(payloadMsg, payloadToSend);

  const char* sensors = "sensors/";
  const uint totalLength = sizeof(char) * (strlen(sensors) + savedUuid.length() + strlen(type) + 2);
  char* topic = (char*)malloc(totalLength);
  strlcpy(topic, sensors, totalLength);
  strlcat(topic, savedUuid.c_str(), totalLength);
  strlcat(topic, "/", totalLength);
  strlcat(topic, type, totalLength);

  Serial.println("notifyValue - publishing topic=" + String(topic));
  mqttClient.publish(topic, payloadToSend);
}

/*
* registerServer function 
* returns an uint:
*  0 => registered or already registered successfully
*  1 => cannot register, because http status code is not 200 (ok) or 209 (already registered)
*  2 => register success, but cannot save the response UUID in preferences
*  3 => cannot deserialize register JSON response payload (probably malformed or too big)
*/
uint registerServer() {
  HTTPClient http;
  # if SSL==true
  client.setCACert(ca_cert);
  # endif

  char* registerUrl = getRegisterUrl();
  Serial.print("registerServer - RegisterUrl: ");
  Serial.println(registerUrl);

  http.begin(client, registerUrl);
  http.addHeader("Content-Type", "application/json; charset=utf-8");

  macAddress = WiFi.macAddress();
  String features = "[";
  features += "{\"type\": \"sensor\",\"name\": \"airquality\",\"enable\": true,\"priority\": 1,\"unit\": \"-\"}";
  features += "]";

 String registerPayload = "{\"mac\": \"" + WiFi.macAddress() + 
    "\",\"manufacturer\": \"" + MANUFACTURER +
    "\",\"model\": \"" + MODEL +
    "\",\"apiToken\": \"" + API_TOKEN +   
    "\",\"features\": " + features + "}";

  Serial.print("registerServer - Register with payload: ");
  Serial.println(registerPayload);
  const int httpResponseCode = http.POST(registerPayload);
  if (httpResponseCode <= 0) {
    Serial.print("registerServer - Error on sending POST with httpResponseCode = ");
    Serial.println(httpResponseCode);
    http.end();

    Serial.println("registerServer - Retrying in 60 seconds...");
    delay(60000);
    // call again registerServer() recursively after the delay
    return registerServer();
  }

  Serial.print("registerServer - httpResponseCode = ");
  Serial.println(httpResponseCode);

  if (httpResponseCode != HTTP_CODE_OK && httpResponseCode != HTTP_CODE_CONFLICT) {
    Serial.println("registerServer - Bad httpResponseCode! Cannot register this device");
    return 1;
  }

  if (httpResponseCode == HTTP_CODE_OK) {
    Serial.println("registerServer - HTTP_CODE_OK");
    StaticJsonDocument<2048> staticDoc;
    DeserializationError err = deserializeJson(staticDoc, http.getStream());
    // There is no need to check for specific reasons,
    // because err evaluates to true/false in this case,
    // as recommended by the developer of ArduinoJson
    if (!err) {
      Serial.println("registerServer - Deserialization succeeded!");
      serializeJsonPretty(staticDoc, Serial);
      const char* uuidValue = staticDoc["uuid"];
      const char* macValue = staticDoc["mac"];
      const char* manufacturerValue = staticDoc["manufacturer"];
      const char* modelValue = staticDoc["model"];
      Serial.print("registerServer - uuidValue: ");
      Serial.println(uuidValue);
      Serial.print("registerServer - macValue: ");
      Serial.println(macValue);
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
      //   return 4;
      // }
      // if (!macAddress.equals(macValue) || nameValue != NAME || manufacturerValue != MANUFACTURER || modelValue != MODEL) {
      //   Serial.println("--- ERROR : Request and response data don't match ---");
      //   return 5;
      // }

      preferences.begin("device", false); 
      String uuidStr = String(uuidValue);
      size_t len = preferences.putString("uuid", uuidStr);
      preferences.end();
      if (len != strlen(uuidValue)) {
        Serial.println("************* ERROR **************");
        Serial.println("registerServer - Cannot SAVE UUID in Preferences");
        Serial.println("**********************************");
        return 2;
      }
    } else {
      Serial.println("registerServer - cannot deserialize register JSON payload");
      return 3;
    }
  } else if (httpResponseCode == HTTP_CODE_CONFLICT) {
    // this is not an error, it appears every reboot after the first registration
    Serial.println("registerServer - HTTP_CODE_CONFLICT - already registered");
  }
  return 0; // OK - registered without errors
}

void readSensorValue() {
  Serial.println("readSensorValue - called");
  int newValue = -1; // initial uninitialized value
  int quality = sensor.slope();
  Serial.print("readSensorValue - sensor value: ");
  Serial.println(sensor.getValue());
  if (quality == AirQualitySensor::FORCE_SIGNAL) {
      Serial.println("readSensorValue - Very high pollution! Force signal active.");
      newValue = AirQualitySensor::FORCE_SIGNAL;
  } else if (quality == AirQualitySensor::HIGH_POLLUTION) {
      Serial.println("readSensorValue - High pollution!");
      newValue = AirQualitySensor::HIGH_POLLUTION;
  } else if (quality == AirQualitySensor::LOW_POLLUTION) {
      Serial.println("readSensorValue - Low pollution!");
      newValue = AirQualitySensor::LOW_POLLUTION;
  } else if (quality == AirQualitySensor::FRESH_AIR) {
      Serial.println("readSensorValue - Fresh air.");
      newValue = AirQualitySensor::FRESH_AIR;
  }

  // notifiy only if the value is changed to prevent useless notifications
  if (currentValue != newValue) {
    currentValue = newValue;
    Serial.println("readSensorValue - notifying value...");
    notifyValue("airquality", currentValue);
  }
}

void initSensor() {
  Serial.println("initSensor - called");
  if (sensor.init()) {
    Serial.println("initSensor - sensor initialized successfully!");
  } else {
    Serial.println("initSensor - cannot initialize sensor");
  }
}

void setup() {
  Serial.begin(115200);
  // To be able to connect Serial monitor after reset or power up and before first print out. Do not wait for an attached Serial Monitor!
  delay(4000);

  # if SSL==true
  Serial.println("setup - Running with SSL enabled");
  # else 
  Serial.println("setup - Running WITHOUT SSL");
  # endif

  Serial.println("--------------- WiFi ----------------");
  WiFi.begin(ssid, password);
  while (WiFi.status() != WL_CONNECTED) {
    delay(500);
    Serial.print(".");
  }

  Serial.println("setup - WiFi connected!");
  Serial.print("setup - IP address: ");
  Serial.println(WiFi.localIP());
  Serial.print("setup - MAC address: ");
  Serial.println(WiFi.macAddress());

  Serial.println("setup - Registering this device...");
  uint result = registerServer();
  if (result == 0) {
    Serial.println("setup - Starting 'notify*' functions...");
    setTime(0,0,0,1,1,21); // set time to Saturday 00:00:00am Jan 1 2021
    Alarm.timerRepeat(15, readSensorValue);
  } else {
    Serial.println("setup - registerServer() returned error code, cannot continue");
    return;
  }

  Serial.println("setup - Getting saved UUID from preferences...");
  preferences.begin("device", false); 
  savedUuid = preferences.getString("uuid", "");
  preferences.end();

  if (savedUuid.equals("")) {
    Serial.println("************* ERROR **************");
    Serial.println("setup - Cannot read saved UUID from Preferences");
    Serial.println("**********************************");
    return;
  }

  initSensor();

  delay(1500);
}

void loop() {
  // if 'savedUuid' is not defined, you cannot use this device
  if (savedUuid == NULL || savedUuid.length() == 0) {
    Serial.println("loop - savedUuid NOT FOUND, cannot continue...");
    delay(60000);
    return;
  }

  if (!mqttClient.connected()) {
    Serial.println("loop - RECONNECTING...");
    reconnect();
  }

  mqttClient.loop();

  Alarm.delay(1000);
}