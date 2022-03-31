# Air Conditioner 

Install global go stuff

go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest

## Install air to watch changes and rebuild

https://github.com/cosmtrek/air

Install it with: `curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin`
Run it with: `air`

```
cd api-server
go mod download
```


## Re-generate gRPC from proto files

From air-conditioner root folder run:

```
protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
api-server/device/device.proto

protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
api-server/register/register.proto

protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
api-devices/device/device.proto

protoc --go_out=. --go_opt=paths=source_relative \
--go-grpc_out=. --go-grpc_opt=paths=source_relative \
api-devices/register/register.proto
```


## RabbitMQ
Run rabbitmq via Docker:
`docker run -d --name rabbitmq -p 8080:15672 -p 5672:5672 rabbitmq:3-management`

Access with:
- user: guest
- password: guest


## Mosquitto
cd git/air-conditioner

docker pull eclipse-mosquitto

docker run -it --name mosquitto -p 1883:1883 -p 9001:9001 --rm -v $PWD/mosquitto.conf:/mosquitto/config/mosquitto.conf -v /mosquitto/data -v /mosquitto/log eclipse-mosquitto

mosquitto_sub -t devices/+/onoff
mosquitto_pub -m "{\"uuid\": \"uuid1\",\"profileToken\": \"profiletoken-1\",\"on\": false}" -t devices/uid1/onoff


### Security

1. Create Certificate Authority

  openssl genrsa -out ca.key 2048
  openssl req -x509 -new -key ca.key -days 3650 -out ca.crt

  ```
  -----
  Country Name (2 letter code) []:IT
  State or Province Name (full name) []:MILANO
  Locality Name (eg, city) []:MILANO
  Organization Name (eg, company) []:CERTIFICATE AUTHORITY
  Organizational Unit Name (eg, section) []:
  Common Name (eg, fully qualified host name) []:192.168.1.71
  Email Address []:<YOUR_EMAIL_ADDRESS>
  ```

2. Create certificate for the Mosquitto server

  openssl genrsa -out server.key 2048
  openssl req -new -key server.key -out server.csr

  ```
  -----
  Country Name (2 letter code) []:IT
  State or Province Name (full name) []:MILANO
  Locality Name (eg, city) []:MILANO
  Organization Name (eg, company) []:MQTT SERVER
  Organizational Unit Name (eg, section) []:MQTT
  Common Name (eg, fully qualified host name) []:192.168.1.71
  Email Address []:<YOUR_EMAIL_ADDRESS>
  
  Please enter the following 'extra' attributes
  to be sent with your certificate request
  A challenge password []:.
  ```

  openssl x509 -req -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt -days 3650

3. Kill mosquitto and resta it with this command:

  docker run -it --name mosquitto -p 1883:1883 -p 9001:9001 --rm -v $PWD/mosquitto-certs/server/certificates:/mosquitto/certificates -v $PWD/mosquitto-certs/mosquitto.conf:/mosquitto/config/mosquitto.conf -v /mosquitto/data -v /mosquitto/log eclipse-mosquitto


## Mondogb

Install mongodb in docker with:

```
docker run -d --name mongodb -v ~/mongodb:/data/db -p 27017:27017 mongo:latest
```


## Docker compose 

```
cd api-server
docker build --tag ks89/ac-api-server .
docker push ks89/ac-api-server
cd ..
cd api-devices
docker build --tag ks89/ac-api-devices .
docker push ks89/ac-api-devices
cd ..
cd react-gui
docker build --tag ks89/ac-gui .
docker push ks89/ac-gui
cd ..
docker-compose up
```

Visit `http://localhost:8085`


## Kubernetes

Follow the tutorial to install and configure Hetzner Cloud in `docs/hetzner-install.md`


## APIS

APIs /api/v1/:

- GET homes/
- POST homes/
- PUT homes/:id
- DELETE homes/:id
- GET homes/:id/rooms/
- POST homes/:id/rooms/
- PUT homes/:id/rooms/:id
- DELETE homes/:id/rooms/:id

- GET airconditioner/
- POST airconditioner/
- PUT airconditioner/:id
- DELETE airconditioner/:id


DB collections

- Home
  - Id
  - Name
  - Location
  - Room[{
    - Id
    - Name
    - Floor
    - AirConditioner[]
  }...]

- AirConditioner
  - Id
  - Name
  - Manufacturer
  - Model
  - Status{
    - On
    - Mode (heat/cold/auto/dry)
    - TargetTemperature
    - Fan{
      - Mode
      - Speed
      - TODO
    }
  }
  


