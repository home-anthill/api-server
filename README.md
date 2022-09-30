# Air Conditioner 

Personal project to control a Beko air conditioner using an ESP32 with an IR emitter.
On server-side, I'm using a Kubernetes cluster with a simple microservices architecture.

## Architecture

<br/>
<img src="https://raw.githubusercontent.com/Ks89/air-conditioner/master/docs/diagrams/air-condirioner-architecture.png" alt="@ks89/air-conditioner">
<br/>

## Install

### Server

To install and configure please follow this official tutorial `docs/hetzner-install.md`

### Device

At the moment, the only supported device is ESP32 S2 (DevKit-C)

1. Rename `secrets.h.template` to `secrets.h`
2. Modify `secrets.h` with your configuration. The most important values are wi-fi credentials, api-token, server-url and mqtt-url.

    ```
    #define SECRET_SSID "WIFI_SSID"
    #define SECRET_PASS "WIFI_PASSWORD"
    ...
    #define API_TOKEN "API_TOKEN_FROM_PROFILE_PAGE"
    #define SERVER_URL "https://SERVER.COM/api/register"
    #define MQTT_URL "MQTT-SERVER.COM"
    ```
   
   To generate the API_TOKEN you have to login to the gui via `https://<SERVER.COM>` with GitHub, then click on the profile icon (upper right corner of the page) to open the profile page.
   In that page, you can re-generate the api-token and copy it in `secrets.h`.
5. Build and flash the firmware via Arduino IDE


## :fire: Releases :fire:

- ??/??/2022 - 1.0.0-alpha.5 - [HERE](https://github.com/Ks89/air-conditioner/releases)
- 08/26/2022 - 1.0.0-alpha.4 - [HERE](https://github.com/Ks89/air-conditioner/releases)
- 05/25/2022 - 1.0.0-alpha.3 - [HERE](https://github.com/Ks89/air-conditioner/releases)
- 05/18/2022 - 1.0.0-alpha.2 - [HERE](https://github.com/Ks89/air-conditioner/releases)
- 05/15/2022 - 1.0.0-alpha.1 - [HERE](https://github.com/Ks89/air-conditioner/releases)


<br/>

# :copyright: License :copyright:

The MIT License (MIT)

Copyright (c) 2021-2022 Stefano Cappa (Ks89)

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NON INFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.

<br/>
