package main

import (
	"encoding/json"
	"fmt"
	"github.com/gorilla/websocket"
	"log"
	"os"
	"os/signal"
)

type GPS struct {
	Lat string `json:"lat"`
	Lng string `json:"lng"`
}

type Truck struct {
	LicenseCar string `json:"license_car"`
	Available  bool   `json:"available"`
}

//type Sensor struct {
//	Fuel   string `json:"fuel"`
//	Weight string `json:"weight"`
//}

type Data struct {
	Name string `json:"name"`
	GPS  GPS    `json:"gps"`
	//Sensor Sensor `json:"sensor"`
}

func main() {

	listData := []Truck{
		{LicenseCar: "59H-21311", Available: false},
		{LicenseCar: "59H-21312", Available: true},
		{LicenseCar: "59H-21313", Available: true},
	}

	// Create WebSocket connection
	addr := "103.124.93.27"
	port := "8080"
	url := "ws://" + addr + ":" + port + "/websocket"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket:", err)
	}
	defer conn.Close()

	checkTruck := make(map[string]bool)

	dataCh := make(chan Data)

	// Start goroutine to read messages from WebSocket
	go func() {
		for {
			_, msg, err := conn.ReadMessage()
			if err != nil {
				log.Println("Error reading message from WebSocket:", err)
				return
			}

			var data Data
			if err = json.Unmarshal(msg, &data); err != nil {
				log.Println("Error parsing JSON:", err)
				continue
			}

			//// vòng lặp qua danh sách xe hiện có
			for i := 0; i < len(listData); i++ {

				// nếu tên xe lấy từ socket không có trong danh sách xe của chủ xe thì bỏ qua
				if data.Name != listData[i].LicenseCar {
					continue
				}

				// check ok trong checkTruck
				// nếu ok trả về false thì append vào list truck và thêm vào map
				if _, ok := checkTruck[data.Name]; !ok {
					checkTruck[data.Name] = true
					dataCh <- data
				}
			}
		}
	}()

	go func() {
		for {
			filterData := <-dataCh
			fmt.Println("filterData: ", filterData)
			log.Println("Name:", filterData.Name)
			log.Println("GPS Lat:", filterData.GPS.Lat)
			log.Println("GPS Lng:", filterData.GPS.Lng)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	<-interrupt
	log.Println("Shutting down...")
}
