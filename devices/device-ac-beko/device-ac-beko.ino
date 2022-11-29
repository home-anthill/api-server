// include the WiFi library and HTTPClient
#include <WiFi.h>
#include <HTTPClient.h>
// include json library (https://github.com/bblanchon/ArduinoJson)
#include <ArduinoJson.h>
// include MQTT library
#include <PubSubClient.h>
// eeprom lib has been deprecated for esp32, the recommended way is to use Preferences
#include <Preferences.h>

// include libraries
// - IRremoteESP8266: https://github.com/crankyoldgit/IRremoteESP8266
#include <IRremoteESP8266.h>
#include <IRsend.h>
// Import the specific implementation to use COOLIX protocol to control Beko ACs
#include <ir_Coolix.h>


#include "secrets.h"


// ------------------------------------------------------
// ------------------ IRremoteESP8266 -------------------
// GPIO pin to use to send IR signals
#define IR_SEND_PIN 4
// ------------------------------------------------------
// ---------------- COOLIX protocol ---------------------
#define SEND_COOLIX
// Temoerature ranges
#define TEMP_MIN kCoolixTempMin // 17
#define TEMP_MAX kCoolixTempMax // 30
// Mode possibile values (defined in ir_Coolix.h)
#define MODE_COOL kCoolixCool // 0
#define MODE_DRY kCoolixDry // 1
#define MODE_AUTO kCoolixAuto // 2
#define MODE_HEAT kCoolixHeat // 3
#define MODE_FAN kCoolixFan // 4
// Fan values (defined in ir_Coolix.h)
#define FAN_AUTO0 kCoolixFanAuto0 // 0
#define FAN_MAX kCoolixFanMax // 1
#define FAN_MED kCoolixFanMed // 2
#define FAN_MIN kCoolixFanMin // 4
#define FAN_AUTO kCoolixFanAuto // 5
// global initial state
struct state {
  bool powerStatus = false;
  uint8_t temperature = TEMP_MAX;
  uint8_t operation = MODE_COOL; // mode (heat, cold, ...)
  uint8_t fan = FAN_AUTO0;
};
state acState;
 // Create a A/C object using GPIO to sending messages with
IRCoolixAC ac(IR_SEND_PIN);
// ------------------------------------------------------
// ------------------------------------------------------


void callbackMqtt(char* topic, byte* payload, unsigned int length);


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
PubSubClient mqttClient(mqttUrl, mqttPort, callbackMqtt, client);


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

void subscribeDevices(const char* command) {
  Serial.print("subscribeDevices - creating topic based on command...");
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
    Serial.println("reconnect - attempting MQTT connection...");
    mqttClient.setBufferSize(4096);
    
    // here you can use the version with `connect(const char* id, const char* user, const char* pass)` with authentication
    const char* idClient = savedUuid.c_str();
    Serial.print("reconnect - connecting to MQTT with client id = ");
    Serial.println(idClient);
    
    if (mqttClient.connect(idClient)) {
      Serial.print("reconnect - connected and subscribing with savedUuid: ");
      Serial.println(savedUuid);
      // subscribe
      subscribeDevices("/values");
    } else {
      Serial.print("reconnect - failed, rc=");
      Serial.print(mqttClient.state());
      Serial.println(" - try again in 5 seconds");
      // Wait 5 seconds before retrying
      delay(5000);
    }
  }
}

void sendIRSignal() {
  Serial.println("sendIRSignal - Sending value via IR...");
  ac.send();
  Serial.println("sendIRSignal - Value sent successfully!");
}

void callbackMqtt(char* topic, byte* payload, unsigned int length) {
  Serial.println("callbackMqtt - called");
 
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
      Serial.println("callbackMqtt - Setting On");
      ac.on();
    } else if (on == 0) {
      Serial.println("callbackMqtt - Setting Off");
      ac.off();
      sendIRSignal();
      Serial.println("--------------------------");
      return;
    }
  }
  if(doc["temperature"] != nullptr) {
    const int temperature = doc["temperature"];
    Serial.print("callbackMqtt - temperature: ");
    Serial.println(temperature);

    if (temperature < TEMP_MIN || temperature > TEMP_MAX) {
      Serial.print("callbackMqtt - Cannot set value, because temperature is out of range. Temperature must be >= ");
      Serial.print(TEMP_MIN);
      Serial.print(" and <= ");
      Serial.print(TEMP_MAX);
      Serial.print("\n");
      return;
    }
    Serial.println("callbackMqtt - Setting temperature");
    ac.setTemp(temperature);
  }
  if(doc["mode"] != nullptr) {
    const int mode = doc["mode"];
    Serial.print("callbackMqtt - mode: ");
    Serial.println(mode);
    switch(mode) {
      case 1:
       Serial.println("callbackMqtt - Setting mode to Cool");
        ac.setMode(MODE_COOL);
        break;
      case 2:
        Serial.println("callbackMqtt - Setting mode to Auto");
        ac.setMode(MODE_AUTO);
        break;
      case 3:
        Serial.println("callbackMqtt - Setting mode to Heat");
        ac.setMode(MODE_HEAT);
        break;
      case 4:
        Serial.println("callbackMqtt - Setting mode to Fan");
        ac.setMode(MODE_FAN);
        break;
      case 5:
        Serial.println("callbackMqtt - Setting mode to Dry");
        ac.setMode(MODE_DRY);
        break;
      default:
        Serial.println("callbackMqtt - Cannot set mode. Unsupported value!");
        break;
    }
  }
  if(doc["fanSpeed"] != nullptr) {
    const int fanSpeed = doc["fanSpeed"];
    Serial.print("callbackMqtt - fanSpeed: ");
    Serial.println(fanSpeed);
    switch(fanSpeed) {
      case 1:
        Serial.println("callbackMqtt - Setting fan speed to Min");
        ac.setFan(FAN_MIN);
        break;
      case 2:
        Serial.println("callbackMqtt - Setting fan speed to Med");
        ac.setFan(FAN_MED);
        break;
      case 3:
        Serial.println("callbackMqtt - Setting fan speed to Max");
        ac.setFan(FAN_MAX);
        break;
      case 4:
        Serial.println("callbackMqtt - Setting fan speed to Auto");
        ac.setFan(FAN_AUTO);
        break;
      case 5:
        Serial.println("callbackMqtt - Setting fan speed to Auto0");
        ac.setFan(FAN_AUTO0);
        break;
      default:
        Serial.println("callbackMqtt - Cannot set fan speed. Unsupported fan value!");
        break;
    }
  }
  sendIRSignal();
  Serial.println("--------------------------");
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
  features += "{\"type\": \"controller\",\"name\": \"ac-beko\",\"enable\": true,\"priority\": 1,\"unit\": \"-\"}";
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

void setup() {
  // Init serial port
  Serial.begin(115200);
  // To be able to connect Serial monitor after reset or power up and before first print out.
  // Do not wait for an attached Serial Monitor!
  delay(3000);
  // disable ESP32 Devkit-C built-in LED
  pinMode (LED_BUILTIN, OUTPUT);
  digitalWrite(LED_BUILTIN, LOW);

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
    Serial.println("setup - registered");
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

  // Run the calibration to calculate uSec timing offsets for this platform.
  // This will produce a 65ms IR signal pulse at 38kHz.
  // Only ever needs to be run once per object instantiation, if at all.
  ac.calibrate();
  delay(1000);
  // Start AC
  ac.begin();
  
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

  delay(1000);
}