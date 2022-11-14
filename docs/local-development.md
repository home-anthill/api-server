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


## 3. Install NodeJS


Install NodeJS LTS from [HERE](https://nodejs.org/)


## 4. Install and run Docker Desktop


Install Docker Desktop from [HERE](https://www.docker.com/products/docker-desktop/)


## 5. Deploy local docker containers


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


## 6. Create GitHub oAuth2 application


```bash
cp devices/api-server/.env_template .env
```

You have to update `OAUTH2_CLIENTID` and `OAUTH2_SECRETID` properties in `.env` file.
These 2 values are the clientID and secretID of your github oAuth2 application, so you need to follow these steps:
1. create an [oAuth2 app on Github](https://docs.github.com/en/developers/apps/building-oauth-apps/creating-an-oauth-app)
2. go to the configuration page of your oAuth2 app and copy the Client ID (**this is the OAUTH2_CLIENTID value**)
3. generate a new client secret and copy it to the `.env` file (**this is the OAUTH2_SECRETID value**)
4. fill the `Homepage URL` input field: `http://localhost:8082`
5. fill the `Authorization callback URL` input field: `http://localhost:8082/api/callback`
6. save the oAuth2 app


## 7. Run all microservices


Open every microservice in a terminal tab (or multiple windows)

1. api-server

```bash
cd devices/api-server
make run
```

2. api-devices

```bash
cd devices/api-devices
make run
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

6. gui

```bash
cd gui
npm run build
```

7. login to the webapp with your GitHub account

If everything is up and running, **you should be able to access at `http://localhost:8082`** from your favourite browser.
From `http://localhost:8082` **login with the GitHub account used to create the oAuth2 application**.
If you'll login successfully you'll be redirected to the main app page.


## 8. Fill database with some data


At this point, you should be able to login to the app, so the DB has a valid profile inside.
However, you don't have any other data.
You can navigate across the webapp to add homes, rooms and so on, but I prefer to show how to insert data manually via APIs using the free [Postman](https://www.postman.com/) desktop app.

### Postman


1. click on `cookies` on the bottom bar
2. enable the "Cookies interceptor" on Domains = `localhost`
<img src="https://raw.githubusercontent.com/Ks89/air-conditioner/master/docs/images/postman-cookies-interceptor.png" alt="Postman cookies interceptor">

3. Download and import in Postman this file `docs/postman-collections/postman_collection.json`

### JWT


1. From your browser, login via GitHub at `http://localhost:8082`
2. Open the "Developer tools" and copy JWT `token` value (standard format `xxxx.xxxx.xxxx`) from "Local Storage" (in Chrome, you can find "Local Storage" under the "Application" tab).
3. Paste this JWT into `authToken` value of collection `Variables` in Postman

<img src="https://raw.githubusercontent.com/Ks89/air-conditioner/master/docs/images/postman-variables-jwt.png" alt="Postman collection variable authToken">

4. Select `getProfile` request (because it requires JWT authentication) from the collection `api-server` and click on the `Send` button. The response should be something like this:
```
{
    "profile": {
        "id": "<YOUR PROFILE MONGODB OBJECTID>",
        "github": {
            "login": "<YOUR GITHUB NICKNAME>",
            "name": "<YOUR GITHUB NAME>",
            "email": "<YOUR GITHUB EMAIL>",
            "avatarURL": ""<YOUR GITHUB AVATAR URL>"
        }
    }
}
```

5. You can try all other requests, but be sure to update path and query parameters with your object ids.
For example, **to get the `apiToken` (required in the next steps) you have to call `regenApiToken` changing the fake profile id from the path param with your profile id.**
You can get your profile id from the response of step 4 (above) and update the path in this way:
```
localhost:8082/api/profiles/<YOUR PROFILE MONGODB OBJECTID>/tokens
```
**The response of `regenApiToken` contains the re-generated `apiToken`**. This token changes every time you call the API and the previous value won't be valid anymore.


## 9. Prepare ESP32 boards with wiring and electrical parts


**TODO add a tutorial and some photos**


## 10. Flash and power on devices


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

// EXTREMELY IMPORTANT!!!
// apiToken (check "section '8. Fill database with some data' > 'Postman' > point 5" in this file above)
#define API_TOKEN "API_TOKEN_FROM_PROFILE_PAGE"

#define SERVER_URL "http://192.168.1.7/api/register"  // add your local IP (for example 192.168.1.7)
#define MQTT_URL "192.168.1.7"        // your local IP (for example 192.168.1.7)
```

4. Open `devices/device`, build and flash the firmware on a ESP32S2 DevKit-C boad
5. Open `sensors/sensor`, build and flash the firmware on a ESP32S2 DevKit-C boad