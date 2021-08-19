Install global go stuff

go install golang.org/x/tools/go/analysis/passes/shadow/cmd/shadow@latest

```
cd server
go mod tidy
```


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
  


