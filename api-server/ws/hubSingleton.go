package ws

import (
  "fmt"
  "sync"
)

var once sync.Once

var singleInstance *Hub

func GetInstance() *Hub {
  if singleInstance == nil {
    once.Do(func() {
      fmt.Println("Creating Single Instance Now")
      singleInstance = newHub()
    })
  } else {
    fmt.Println("Single Instance already created")
  }
  return singleInstance
}
