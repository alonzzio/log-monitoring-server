/*
	Log Monitoring System ( LMS )
	This program is indented to demonstrate the functionalities only.
	Not fully focused on complete error handling in place.
	However, I'll try my best to cover error handling in place

	Copy right: Ratheesh Alon Rajan
*/

package main

import (
	"fmt"
	"github.com/alonzzio/log-monitoring-server/internal/config"
	"log"
)

// app holds application wide configs
var app config.AppConfig

func init() {
	log.Println("Log monitoring Server starting up...")
}

func main() {
	err := run()
	if err != nil {
		log.Fatal(err)
	}

	var conn *config.Conn
	conn, err = newConn()
	if err != nil {
		log.Fatal(err)
	}

	//set conn to App
	app.Conn = conn

	fmt.Println("end")
}
