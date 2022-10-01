# Local development setup


## 1. Install Go


1. Install Go from [HERE](https://go.dev/)
2. Install global go stuff `go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest`
3. Install [air](https://github.com/cosmtrek/air) to watch changes and auto-rebuild

```bash
curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin
```


## 2. Install Rust


Install Rust from [HERE](https://www.rust-lang.org/)


## 3. Install and run Docker Desktop


Install Docker Desktop from [HERE](https://www.docker.com/products/docker-desktop/)


## 4. Deploy local docker containers


1. Mosquitto

```bash
# from the root folder of this repo
docker pull eclipse-mosquitto
docker run -it --name mosquitto -p 1883:1883 -p 9001:9001 --rm -v $PWD/mosquitto/mosquitto-no-security.conf:/mosquitto/config/mosquitto.conf -v /mosquitto/data -v /mosquitto/log eclipse-mosquitto
```
**Don't close this terminal window!**

2. RabbitMQ

```bash
docker pull rabbitmq:management
docker run -d --name rabbitmq --hostname my-rabbit -p 8080:15672 -p 5672:5672 rabbitmq:management
```

If you want you can access to the UI at `http://locahost:8080` and login with:
```
user: guest
password: guest
```

3. MongoDB

```bash
docker run -d --name mongodb -v ~/mongodb:/data/db -p 27017:27017 mongo:6
```


## 5. Run all microservices


Open every microservice in a terminal tab (or multiple windows)

1. api-server

```bash
cd devices/api-server
make air
```

2. api-devices

```bash
cd devices/api-devices
make air
```

3. sensor register

```bash
cd sensors/register
cargo run
```

4. producer

```bash
cd sensors/producer
cargo run
```

5. consumer

```bash
cd sensors/consumer
cargo run
```

## 6. Fill database with some data


**TODO add a tutorial to do that calling APIS via Postman**

**from this you should copy the apiToken of your profile**


## 7. Prepare ESP32 boards with wiring and electrical parts

**TODO add a tutorial and some photos**


## 8. Flash and power on devices


1. Configure [Arduino IDE 2.x](https://www.arduino.cc/en/software) to build and flash ESP32S2 dev-kitC firmwares. You need the `esp32` board in `Board Manager` as described in [the official tutorial](https://espressif-docs.readthedocs-hosted.com/projects/arduino-esp32/en/latest/installing.html).
Then try to build and flash an an official example to see if everything is ok!
2. From Arduino IDE install these libraries from `Library Manager` tab:
- `Arduino Unified Sensor` by Adafruit (version `1.1.6`)
- `ArduinoJson` by Benoit Blanchon (version `6.19.4`)
- `DHT sensor library` by Adafruit (version `1.4.4`)
- `HttpClient` by Adrian McEwen (version `2.2.0`)
- `IRremote` by shirriff, z370, ArminJo... (version `3.9.0`)
- `PubSubClient` by Nick O'Leary (version `2.8`)
- `Time` by Michael Margolis (version `1.6.1`)
- `TimeAlarms` by Michael Margolis (version `1.5`)
3. Prepare `secrets.h` files:

```bash
cp devices/device/secrets.h.template devices/device/secrets.h
cp sensors/sensor/secrets.h.template sensors/sensor/secrets.h
```

4. Update `secrets.h` files in this way:

First check the IP address of your pc (based on your OS):

```bash
ip a
# or
ifconfig
# or
ipconfig /a
```

You'll have something like `192.168.1.???`, for example `192.168.1.7`

```cpp
// update the content of this file and rename it to 'secrets.h'

#define SECRET_SSID "WIFI_SSID"       // this is the name of your WiFi network
#define SECRET_PASS "WIFI_PASSWORD"   // this is the password of your WiFi network

#define MANUFACTURER "MANUFACTURER"   // choose a manufacturer name
#define MODEL "MODEL"                 // choose a model name
#define API_TOKEN "API_TOKEN_FROM_PROFILE_PAGE"   // apiToken (see section above)

#define SERVER_URL "http://192.168.1.7/api/register"  // add your local IP (for example 192.168.1.7)
#define MQTT_URL "192.168.1.7"        // your local IP (for example 192.168.1.7)
```

4. Open `devices/device`, build and flash the firmware on a ESP32S2 DevKit-C boad
5. Open `sensors/sensor`, build and flash the firmware on a ESP32S2 DevKit-C boad