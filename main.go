package main

import (
	"fmt"
	"os"
	"time"

	"github.com/itsMeDacarvalho/ubiwhere-challenge/database"
	"github.com/itsMeDacarvalho/ubiwhere-challenge/toolset"

	"github.com/boltdb/bolt"
)

func main() {
	//Setup a new database passing name as argument
	db, err := database.SetupDB("ubiDB")

	// Create new section into log file
	toolset.WriteToLog("+----------------------------------------+\n")
	toolset.WriteToLog(fmt.Sprintf("| Session: %s - %s \t\t | \n", toolset.GetFormatedDate(), toolset.GetFormatedTime()))
	toolset.WriteToLog("+----------------------------------------+\n")
	toolset.WriteToLog(fmt.Sprintf("Time \t\t || Source \t Message\n\n"))

	//Handle error from setupDB
	if err != nil {
		toolset.WriteToLog(fmt.Sprintf("%s \t || [Database] \t Cant perform setup: %v \n", toolset.GetFormatedTime(), err))
	}
	//Sucess creating new database
	toolset.WriteToLog(fmt.Sprintf("%s \t || [Database] \t Init setup performed with success \n", toolset.GetFormatedTime()))

	//Make sure db close at the end of this function
	defer db.Close()

	//Set DB configuration: only created time
	config := database.Config{LastAccessTime: time.Now().Format("06/01/02 15:04:05")}
	err = database.SetConfig(db, config)

	//Handle possible setConfig errors
	if err != nil {
		toolset.WriteToLog(fmt.Sprintf("%s \t || [Database] \t Something went wrong with configuration: %v \n", toolset.GetFormatedTime(), err))
	}

	go scheduleDataOS(db)
	go scheduleSampleData(db)

	// Wait one second to sample the first data
	time.Sleep(time.Second * 1)

	// opt string to get user data
	opt := ""

	// Escape code to clear console
	fmt.Println("\033[H\033[2J")

	for opt != "0" {
		opt, value := toolset.PrintMenu()

		switch opt {
		case "0":
			fmt.Printf("\nProgram is now exiting...\n")
			os.Exit(0)
			break
		case "1":
			data, _ := toolset.GetLastN(db, value)

			// Escape code to clear console
			fmt.Println("\033[H\033[2J")

			// Print data
			toolset.PrintLastN(data, value)
		case "2":
			data, _ := toolset.GetLastN(db, value)
			fmt.Printf("%v\n", value)

			// Escape code to clear console
			fmt.Println("\033[H\033[2J")

			// Print data
			toolset.PrintLastN(data, value)
		case "3":
			// Escape code to clear console
			fmt.Println("\033[H\033[2J")

			toolset.PrintAverage(db, value)
		}
	}

	// select allow concurrent go functions to run before closing the program
	select {}
}

// scheduleDataOS - schecule OS data entry every second
func scheduleDataOS(db *bolt.DB) {
	toolset.WriteToLog(fmt.Sprintf("%s \t || [OS] \t OS data is now being collected every second \n", toolset.GetFormatedTime()))

	// Start new ticker, in order to repeat something every second
	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {
		toolset.StoreDataOS(db)
	}
}

// scheduleDataOS - schecule OS data entry every second
func scheduleSampleData(db *bolt.DB) {
	toolset.WriteToLog(fmt.Sprintf("%s \t || [SAMPLE] \t Sample data from simulator is now being collected every second \n", toolset.GetFormatedTime()))

	// Start new ticker, in order to repeat something every second
	ticker := time.NewTicker(1 * time.Second)

	for range ticker.C {
		toolset.StoreSample(db)
	}
}
