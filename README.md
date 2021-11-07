Install global go stuff

go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest

## Install air to watch changes and rebuild

https://github.com/cosmtrek/air

Install it with: `curl -sSfL https://raw.githubusercontent.com/cosmtrek/air/master/install.sh | sh -s -- -b $(go env GOPATH)/bin`
Run it with: `air`

```
cd server
go mod tidy
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

## Mondogb

Install mongodb in docker with:

```
docker run -d --name mongodb -v ~/mongodb:/data/db -p 27017:27017 mongo:latest
```


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
  


