// include the WiFi library
#include <WiFi101.h>
// include json library (https://github.com/bblanchon/ArduinoJson)
#include <ArduinoJson.h>
#include "secrets.h"

#include <PubSubClient.h>
/*
 * Define macros for input and output pin etc.
 */
#include "PinDefinitionsAndMore.h"

#include <IRremote.h>



// -------------------------------------------------------
// ---------------------- others -------------------------
const int serverPort = 3000;
const int serverIp[4] = {192, 168, 1, 71};
const int serverPortMqtt = 1883;
IPAddress serverMqtt(192, 168, 1, 71);

// -------------------------------------------------------
// -------------------------------------------------------
///////please enter your sensitive data in the Secret tab/arduino_secrets.h
char ssid[] = SECRET_SSID;        // your network SSID (name)
char pass[] = SECRET_PASS;    // your network password (use for WPA, or use as key for WEP)
int keyIndex = 0;            // your network key Index number (needed only for WEP)
int status = WL_IDLE_STATUS;
// if you don't want to use DNS (and reduce your sketch size)
// use the numeric IP instead of the name for the server:
IPAddress server(serverIp[0],serverIp[1],serverIp[2],serverIp[3]);  // numeric IP (no DNS)
//char server[] = "35.206.99.222";    // name address (using DNS)
// Initialize the Ethernet client library
// with the IP address and port of the server
// that you want to connect to (port 80 is default for HTTP):
WiFiClient client;

PubSubClient mqttClient(client);

void reconnect() {
  // Loop until we're reconnected
  while (!mqttClient.connected()) {
    Serial.print("Attempting MQTT connection...");
    // Attempt to connect
    mqttClient.setBufferSize(4096);
    if (mqttClient.connect("arduinoClient")) {
      Serial.println("connected");
      // Once connected, publish an announcement...
      // mqttClient.publish("topic/state","hello world");
      // ... and resubscribe
      mqttClient.subscribe("topic/state");
    } else {
      Serial.print("failed, rc=");
      Serial.print(mqttClient.state());
      Serial.println(" try again in 5 seconds");
      // Wait 5 seconds before retrying
      delay(5000);
    }
  }
}

void callback(char* topic, byte* payload, unsigned int length) {
  Serial.print("Message arrived [");
  Serial.print(topic);
  Serial.print("] ");
  for (int i=0;i<length;i++) {
    Serial.print((char)payload[i]);
  }
  Serial.println("Deserializing json payload...");

// char el[] = "{\"id\":\"dadsd\",\"operation\":\"ON\"}";

  DynamicJsonDocument doc(2048);
  DeserializationError error = deserializeJson(doc, payload);
  if (error) {
    Serial.print(F("deserializeJson() failed: "));
    Serial.println(error.f_str());
    return;
  }

	// JsonVariant docVariant = doc.as<JsonVariant>();

  String id = doc["id"];
  String operation = doc["operation"];

	// JsonVariant jsonVariant = doc.as<JsonVariant>();

  // // Generate the prettified JSON and send it to the Serial port.
  // //
  // serializeJsonPretty(doc, Serial);

  // Serial.println(jsonVariant["id"].c_str());

  // if(!jsonVariant.containsKey("id") || !jsonVariant["id"].is<char*>() ||
	//    !jsonVariant.containsKey("operation") || !jsonVariant["operation"].is<char*>()) {
  //   Serial.println("Bad input message");
	// 	return;
	// }

  // String id = jsonVariant["id"];
	// String operation = jsonVariant["operation"];

  Serial.print("id ");
  Serial.println(id);
  Serial.print("operation ");
  Serial.println(operation);

//   if (operation == "ON") {
//     DynamicJsonDocument jsonResponse(2048);
// 		jsonResponse["id"] = id;
// 		jsonResponse["message"] = "hooray!!!";
// 		String json;
// 		serializeJson(jsonResponse, json);
//     Serial.print("json ");
//     Serial.println(json);
// 		mqttClient.publish(id.c_str(), json.c_str());
//   }
}

void setup() {
   // pinMode(LED_BUILTIN, OUTPUT);

    Serial.begin(9600);
#if defined(__AVR_ATmega32U4__) || defined(SERIAL_USB) || defined(SERIAL_PORT_USBVIRTUAL)  || defined(ARDUINO_attiny3217)
    delay(4000); // To be able to connect Serial monitor after reset or power up and before first print out. Do not wait for an attached Serial Monitor!
#endif
    // Just to know which program is running on my Arduino
    Serial.println(F("START " __FILE__ " from " __DATE__ "\r\nUsing library version " VERSION_IRREMOTE));

    IrSender.begin(IR_SEND_PIN, ENABLE_LED_FEEDBACK); // Specify send pin and enable feedback LED at default feedback LED pin

    Serial.print(F("Ready to send IR signals at pin "));
#if defined(ARDUINO_ARCH_STM32) || defined(ESP8266)
    Serial.println(IR_SEND_PIN_STRING);
#else
    Serial.print(IR_SEND_PIN);
#endif

// --------------------------------------------------------------------
  // ---------------- wifi + json rest api to remote server -------------
  // --------------------------------------------------------------------
  // check for the presence of the shield:
  if (WiFi.status() == WL_NO_SHIELD) {
    Serial.println("WiFi shield not present");
    // don't continue:
    while (true);
  }

  // attempt to connect to WiFi network:
  while (status != WL_CONNECTED) {
    Serial.print("Attempting to connect to SSID: ");
    Serial.println(ssid);
    // Connect to WPA/WPA2 network. Change this line if using open or WEP network:
    status = WiFi.begin(ssid, pass);

    // wait 10 seconds for connection:
    delay(10000);
  }
  Serial.println("Connected to wifi");
  printWiFiStatus();

  // Serial.println("\nStarting connection to server...");
  
  // if (!client.connect(server, serverPort)) {
  //   Serial.println("Connection failed");
  //   return;
  // }

  // Serial.println("Connected!");


  mqttClient.setServer(serverMqtt, serverPortMqtt);
  mqttClient.setCallback(callback);
  
  // // Make a HTTP request:
  // client.println("GET /api/keepAlive HTTP/1.1");
  // // String hostPrefix = "Host: ";
  // // String host = host + serverIp[0] + "." + serverIp[1] + "." + serverIp[2] + "." + serverIp[3];
  // //client.println("Host: 35.206.99.222");
  // client.println("Connection: close");

  // if (client.println() == 0) {
  //   Serial.println("Failed to send request");
  //   return;
  // }

  // // Check HTTP status
  // char status[32] = {0};
  // client.readBytesUntil('\r', status, sizeof(status));
  // if (strcmp(status, "HTTP/1.1 200 OK") != 0) {
  //   Serial.print("Unexpected response: ");
  //   Serial.println(status);
  //   return;
  // }

  // // Skip HTTP headers
  // char endOfHeaders[] = "\r\n\r\n";
  // if (!client.find(endOfHeaders)) {
  //   Serial.println("Invalid response");
  //   return;
  // }


  // // Allocate JsonBuffer
  // // Use arduinojson.org/assistant to compute the capacity.
  // const size_t capacity = JSON_OBJECT_SIZE(3) + JSON_ARRAY_SIZE(2) + 60;
  // DynamicJsonBuffer jsonBuffer(capacity);

  // // Parse JSON object
  // JsonObject& root = jsonBuffer.parseObject(client);
  // if (!root.success()) {
  //   Serial.println("Parsing failed!");
  //   return;
  // }

  // // Extract values
  // Serial.println("Response:");
  // Serial.println(root["year"].as<char*>());
  // Serial.println(root["month"].as<char*>());
  // Serial.println(root["day"].as<char*>());
  // Serial.println(root["hours"].as<char*>());
  // Serial.println(root["minutes"].as<char*>());
  // Serial.println(root["seconds"].as<char*>());
  // Serial.println(root["offset"].as<char*>());

  // Disconnect
  //client.stop();
  // ------------------------------
  // ------------------------------
  // ------------------------------

  // Allow the hardware to sort itself out
  delay(1500);
}

/*
 * NEC address=0xFB0C, command=0x18
 *
 * This is data in byte format.
 * The uint8_t/byte elements contain the number of ticks in 50 us.
 * The uint16_t format contains the (number of ticks * 50) if generated by IRremote,
 * so the uint16_t format has exact the same resolution but requires double space.
 * With the uint16_t format, you are able to modify the timings to meet the standards,
 * e.g. use 560 (instead of 11 * 50) for NEC or use 432 for Panasonic. But in this cases,
 * you better use the timing generation functions e.g. sendNEC() directly.
 */

void loop() {
  const uint8_t NEC_KHZ = 38; // 38kHz carrier frequency for the NEC protocol

  /*
  * Send hand crafted data from RAM
  * The values are NOT multiple of 50, but are taken from the NEC timing definitions
  */
  //Serial.println(F("Send NEC 8 bit address 0xFB04, 0x08 with exact timing (16 bit array format)"));
  //Serial.flush();

  //const uint16_t on[] = { 4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550 }; // Using exact NEC timing

  // const uint16_t off[] = {4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550,4500,4500,4500,550,1650,550,550,550,1650,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,550,550,1650,550,550,550,1650,550,1650,550,1650,550,1650,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,1650,550,550,550,550,550,1650,550,1650,550,1650,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,550,1650,550,1650,550,1650,550,1650,550,1650,550};
  
  // IrSender.sendRaw(off, sizeof(off) / sizeof(off[0]), NEC_KHZ); // Note the approach used to automatically calculate the size of the array.

  if (!mqttClient.connected()) {
    Serial.println("RECONNECTING...");
    reconnect();
  }
  mqttClient.loop();

  delay(1000);
}

// ------------------------------
// ----------- Wifi -------------
// ------------------------------
void printWiFiStatus() {
  // print the SSID of the network you're attached to:
  Serial.print("SSID: ");
  Serial.println(WiFi.SSID());

  // print your WiFi shield's IP address:
  IPAddress ip = WiFi.localIP();
  Serial.print("IP Address: ");
  Serial.println(ip);

  // print the received signal strength:
  long rssi = WiFi.RSSI();
  Serial.print("signal strength (RSSI):");
  Serial.print(rssi);
  Serial.println(" dBm");
}
// ------------------------------
// ------------------------------
// ------------------------------
