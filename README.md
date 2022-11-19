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

1. Write your custom `air-conditioner-server-config/secrets.yaml` config file:

    ```yaml
    wifi_ssid: 'your-wifi-ssid'
    wifi_password: 'your-wifi-password'

    manufacturer: 'ks89'
    api_token: 'API_TOKEN_FROM_PROFILE_PAGE'

    # enable both HTTPS and MQTTS
    # you should change PORTS accordingly
    # https port: 443
    # mqtts port: 8883
    ssl: true

    server_domain: 'your-https-domain.com'
    server_port: '443'
    server_path: '/api/register'

    mqtt_domain: 'your-mqtt-domain.com'
    mqtt_port: 8883
    ```

    To generate the API_TOKEN_FROM_PROFILE_PAGE you have to login to the gui via `https://<SERVER.COM>` with GitHub, then click on the profile icon (upper right corner of the page) to open the profile page.
    In that page, you can re-generate the api-token for your devices.
2. Generate `secrets.h` files for all your devices:

    ```bash
    cd esp32-configurator
    
    python3 -m configurator --model=thl --source=../../air-conditioner-server-config/secrets.yaml --destination=../sensors/sensor-thl
    
    python3 -m configurator --model=light --source=../../air-conditioner-server-config/secrets.yaml --destination=../sensors/sensor-light
    
    python3 -m configurator --model=motion --source=../../air-conditioner-server-config/secrets.yaml --destination=../sensors/sensor-motion

    python3 -m configurator --model=airquality --source=../../air-conditioner-server-config/secrets.yaml --destination=../sensors/sensor-airquality

    python3 -m configurator --model=airpressure --source=../../air-conditioner-server-config/secrets.yaml --destination=../sensors/sensor-airpressure

    python3 -m configurator --model=ac --source=../../air-conditioner-server-config/secrets.yaml --destination=../devices/device
    ```

3. Build and flash firmwares via Arduino IDE


## :fire: Releases :fire:

- ??/??/2022 - 1.0.0-alpha.5 - [HERE](https://github.com/Ks89/air-conditioner/releases)
- 08/26/2022 - 1.0.0-alpha.4 - [HERE](https://github.com/Ks89/air-conditioner/releases)
- 05/25/2022 - 1.0.0-alpha.3 - [HERE](https://github.com/Ks89/air-conditioner/releases)
- 05/18/2022 - 1.0.0-alpha.2 - [HERE](https://github.com/Ks89/air-conditioner/releases)
- 05/15/2022 - 1.0.0-alpha.1 - [HERE](https://github.com/Ks89/air-conditioner/releases)

<br/>

## Local development

To setup this project on your PC to develop and run these microservices, please take a look at `docs/local-development.md`


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
