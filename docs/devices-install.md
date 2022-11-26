# Devices install guide


## 1. Prepare ESP32 boards with wiring and electrical parts

**TODO add a tutorial and some photos**


## 2. Build and flash firmwares


1. Configure [Arduino IDE 2.x](https://www.arduino.cc/en/software) to build and flash ESP32 firmwares. You need the `esp32` board in `Board Manager` as described in [the official tutorial](https://espressif-docs.readthedocs-hosted.com/projects/arduino-esp32/en/latest/installing.html).
Then try to build and flash an an official example to see if everything is ok!

2. From Arduino IDE install these libraries from `Library Manager` tab:
- `ArduinoJson` by Benoit Blanchon (version `6.19.4`)
- `HttpClient` by Adrian McEwen (version `2.2.0`)
- `PubSubClient` by Nick O'Leary (version `2.8`)
- `Time` by Michael Margolis (version `1.6.1`)
- `TimeAlarms` by Michael Margolis (version `1.5`)
- `Arduino Unified Sensor` by Adafruit (version `1.1.6`)
- `DHT sensor library` by Adafruit (version `1.4.4`)
- `IRremoteESP8266` by David Conran, Sebastien Warin, Mark Szabo, Ken Shirriff (version `2.8.4`)

Additionally, you need to manually add other libraries in Arduino folder. To do this, open that folder (`/Users/<YOUR_USERNAME>/Documents/Arduino/libraries/` on macOS) and copy the libraries:
- `Grove - Air quality sensor` by Seeed Studio (latest version on `master branch` or commit hash `58e4c0bb5ce1b0c9b8aa1265e9f726025feb34f0` [FROM GITHUB](https://github.com/Seeed-Studio/Grove_Air_quality_Sensor)). You cannot use the one published on ArduinoIDE Library Manager, because it's outdated.
- `Grove - Digital Light Sensor` by Seeed Studio (latest version on `master branch` or commit hash `69f7175ed1349276364994d1d45041c6e90a129b` [FROM GITHUB](https://github.com/Seeed-Studio/Grove_Digital_Light_Sensor)). You cannot use the one published on ArduinoIDE Library Manager, because it's outdated.
- `Infineon - DPS310 Pressure Sensor` by Infineon (latest version on `master branch` or commit hash `ed02f803fc780cbcab54ed8b35dd3d718f2ebbda` [FROM GITHUB](https://github.com/Infineon/DPS310-Pressure-Sensor)).


3. Create a new file `home-anthill-server-config/secrets.yaml` file with this content

```yaml
wifi_ssid: '<YOUR WIFI SSID>'
wifi_password: '<YOUR WIFI PASSWORD>'

manufacturer: 'ks89'
api_token: '<PROFILE API TOKEN>' # from your local DB

# enable both HTTPS and MQTTS
# you should change PORTS accordingly
# https port: 443
# mqtts port: 8883
ssl: true

server_domain: '<YOUR HTTPS PUBLIC DOMAIN>'
server_port: '443'
server_path: '/api/register'

mqtt_domain: '<YOUR MQTTS PUBLIC DOMAIN>'
mqtt_port: 8883
```

4. Run `esp32-configurator` Python script:

```bash
cd esp32-configurator

python3 -m configurator --model=dht-light --source=../secrets.yaml --destination=../sensors/sensor-dht-light
python3 -m configurator --model=airquality-pir --source=../secrets.yaml --destination=../sensors/sensor-airquality-pir
python3 -m configurator --model=barometer --source=../secrets.yaml --destination=../sensors/sensor-barometer

python3 -m configurator --model=ac --source=../secrets.yaml --destination=../devices/device
```

5. Build and flash firmwares

- Open `devices/device/device.ino` with ArduinoIDE and flash the firmware
- Open `sensors/sensor-dht-light/sensor-dht-light.ino` with ArduinoIDE and flash the firmware
- Open `sensors/sensor-airquality-pir/sensor-airquality-pir.ino` with ArduinoIDE and flash the firmware
- Open `sensors/sensor-barometer/sensor-barometer.ino` with ArduinoIDE and flash the firmware