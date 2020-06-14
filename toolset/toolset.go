package toolset

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"time"
	"ubiwhere-challenge/collector"
	"ubiwhere-challenge/database"
	"ubiwhere-challenge/sim"

	"github.com/boltdb/bolt"
)

// PrintSampleData - Show in terminal well formated SAMPLE table data
func PrintSampleData(db *bolt.DB) error {
	// Temporary struct to decode json
	var tmpSample database.Sample

	err := db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket([]byte("DB")).Bucket([]byte(strings.ToUpper("SAMPLES")))

		//Print header
		fmt.Printf("+--------------------------------------------------------+\n")
		fmt.Printf("| Time \t\t\t S1 \t S2 \t S3 \t S4 \t | \n")
		fmt.Printf("+--------------------------------------------------------+\n")
		fmt.Printf("+--------------------------------------------------------+\n")

		err := data.ForEach(func(k, v []byte) error {
			// Decode json data values
			json.Unmarshal([]byte(v), &tmpSample)

			fmt.Printf("| %s \t %d \t %d \t %d \t %d \t | \n", string(k), tmpSample.Sample1, tmpSample.Sample2, tmpSample.Sample3, tmpSample.Sample4)
			fmt.Printf("+--------------------------------------------------------+\n")

			return nil
		})
		return err
	})

	return err
}

// PrintOSData - Show in terminal well formated OS table data
func PrintOSData(db *bolt.DB) error {
	// Temporary struct to decode json
	var tmpOS database.PerformanceOS

	err := db.View(func(tx *bolt.Tx) error {
		data := tx.Bucket([]byte("DB")).Bucket([]byte("OS"))

		//Print header
		fmt.Printf("+------------------------------------------------------------------------+\n")
		fmt.Printf("| Time \t\t\t CPU (%%) \t Used RAM / Total RAM (Mb)\t | \n")
		fmt.Printf("+------------------------------------------------------------------------+\n")
		fmt.Printf("+------------------------------------------------------------------------+\n")

		err := data.ForEach(func(k, v []byte) error {
			// Decode json data values
			json.Unmarshal([]byte(v), &tmpOS)

			fmt.Printf("| %s \t %.2f %% \t %d / %d Mb (%.2f %%)\t | \n", string(k), tmpOS.CPU, tmpOS.UsedRAM, tmpOS.TotalRAM, ((float64(tmpOS.UsedRAM) / float64(tmpOS.TotalRAM)) * 100))
			fmt.Printf("+------------------------------------------------------------------------+\n")

			return nil
		})
		return err
	})

	return err
}

// StoreDataOS - get info from OS and perform an entry on database
func StoreDataOS(db *bolt.DB) error {
	infoRAM := collector.GetRAM()
	err := database.AddSystemStat(db, collector.GetCPU(), infoRAM[0], infoRAM[1])

	return err
}

// StoreSample - get samples from simulator and perform an entry on database
func StoreSample(db *bolt.DB) error {
	err := database.AddSample(db, sim.GenerateSamples())

	return err
}

// GetLastN - Returns a map with variable as key and metrics as value
func GetLastN(db *bolt.DB, desiredN int) (map[int][]int, error) {
	// Map with type of sample as key to array of last n values
	sampleData := make(map[int][]int)

	// Temporary structs to decode json
	//var tmpOS database.PerformanceOS
	var tmpSamples database.Sample
	var tmpOS database.PerformanceOS

	err := db.View(func(tx *bolt.Tx) error {
		// Get SAMPLES table data
		sampleTable := tx.Bucket([]byte("DB")).Bucket([]byte("SAMPLES"))
		osTable := tx.Bucket([]byte("DB")).Bucket([]byte("OS"))

		//Create new cursor for SAMPLES table and point to last entry
		sampleCursor := sampleTable.Cursor()
		osCursor := osTable.Cursor()

		_, sampleValue := sampleCursor.Last()
		_, osValue := osCursor.Last()

		// Iterate n times over table using cursor and store data into sampleData map
		for i := 0; i < desiredN; i++ {
			if sampleValue == nil || osValue == nil {
				fmt.Printf("[Info] - Desired number of metrics exceed total size of table.\n")
				break
			}
			// Decode value data that contains 4 ints for each sample
			json.Unmarshal([]byte(sampleValue), &tmpSamples)
			json.Unmarshal([]byte(osValue), &tmpOS)

			// Sample data fetched to map
			// Final map to return : [S1 S2 S3 S4 CPU TOTAL_RAM USED_RAM]
			sampleData[desiredN-i] = append(sampleData[desiredN-i], tmpSamples.Sample1)
			sampleData[desiredN-i] = append(sampleData[desiredN-i], tmpSamples.Sample2)
			sampleData[desiredN-i] = append(sampleData[desiredN-i], tmpSamples.Sample3)
			sampleData[desiredN-i] = append(sampleData[desiredN-i], tmpSamples.Sample4)
			sampleData[desiredN-i] = append(sampleData[desiredN-i], int(math.Round(tmpOS.CPU)))
			sampleData[desiredN-i] = append(sampleData[desiredN-i], int(tmpOS.TotalRAM))
			sampleData[desiredN-i] = append(sampleData[desiredN-i], int(tmpOS.UsedRAM))

			// Point cursor to the previous entry
			_, sampleValue = sampleCursor.Prev()
			_, osValue = osCursor.Prev()
		}

		return nil
	})

	if err != nil {
		WriteToLog(fmt.Sprintf("%s \t || [Database] \t Error while getting metrics from SAMPLES table \n", GetFormatedTime()))
	}

	return sampleData, err
}

// PrintLastN - Prints all last n metrics in a well formated way
func PrintLastN(data map[int][]int) {

	//Print header
	fmt.Printf("+----------------------------------------------------------------------------------------+\n")
	fmt.Printf("| # | \t S1 \t S2 \t S3 \t S4 \t | CPU \t\t Used RAM \t Total RAM \t | \n")
	fmt.Printf("+----------------------------------------------------------------------------------------+\n")
	fmt.Printf("+----------------------------------------------------------------------------------------+\n")
	// To store the keys in slice in sorted order
	var keys []int
	for k := range data {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	for _, k := range keys {
		fmt.Printf("| %d | \t %d \t %d \t %d \t %d \t | %d%% \t\t %d Mb \t %d \t\t | \n", k, data[k][0], data[k][1], data[k][2], data[k][3], data[k][4], data[k][6], data[k][5])
		fmt.Printf("+----------------------------------------------------------------------------------------+\n")
	}
}

// PrintOneMoreLastN - Prints all last n metrics choosed by user
func PrintOneMoreLastN(data map[int][]int, codeStr string) {
	fmt.Println(codeStr)
	decodeStr := map[string]string{
		"c": "CPU",
		"r": "RAM",
		"1": "S1",
		"2": "S2",
		"3": "S3",
		"4": "S4",
	}

	codeArray := strings.Split(codeStr, ",")

	metricsNumber := codeArray[0]

	fmt.Printf("Showing %s metrics...\n", metricsNumber)

	//Print header
	fmt.Printf("+----------------------------------------------------------------------------------------+\n|")

	// Get variable name from decode string map and produce the table header
	for i := 1; i < len(codeArray); i++ {
		fmt.Printf("%s \t\t", decodeStr[codeArray[i]])
	}

	fmt.Printf("|\n+----------------------------------------------------------------------------------------+\n")
	fmt.Printf("+----------------------------------------------------------------------------------------+\n")

	// To store the keys in slice in sorted order
	var keys []int
	for k := range data {
		keys = append(keys, k)
	}
	sort.Ints(keys)
	/*
		for _, k := range keys {
			fmt.Printf("| %d | \t %d \t %d \t %d \t %d \t | %d%% \t\t %d Mb \t %d \t\t | \n", k, data[k][0], data[k][1], data[k][2], data[k][3], data[k][4], data[k][6], data[k][5])
			fmt.Printf("+----------------------------------------------------------------------------------------+\n")
		}*/
}

// PrintMenu - Display well formated menu to user and return choosed option
func PrintMenu() (string, string) {
	reader := bufio.NewReader(os.Stdin)
	optStr := ""

	options := map[string]string{
		"CPU":      "c",
		"RAM":      "r",
		"Sample 1": "1",
		"Sample 2": "2",
		"Sample 3": "3",
		"Sample 4": "4",
	}

	fmt.Printf("\n\n+---------------------------------------------------------------+\n")
	fmt.Printf("|\t\t\tUBIWHERE CHALLENGE\t\t\t|\n")
	fmt.Printf("+---------------------------------------------------------------+\n")
	fmt.Printf("| 1 - Get last n metrics for all variables \t\t\t|\n")
	fmt.Printf("| 2 - Get last n metrics for one or more variables \t\t|\n")
	fmt.Printf("| 3 - Get an average of the value of one or more variables \t|\n")
	fmt.Printf("| 0 - Exit \t\t\t\t\t\t\t|\n")
	fmt.Printf("+---------------------------------------------------------------+\n")
	fmt.Printf(">> Option: ")

	// Read string until user ENTER aka newline
	opt, _ := reader.ReadString('\n')

	// Delete return carriage / newline from option readed
	opt = strings.TrimSuffix(opt, "\n")

	switch opt {
	case "1":
		fmt.Printf(">> How many metrics: ")
		n, _ := reader.ReadString('\n')
		n = strings.TrimSuffix(n, "\n")
		return opt, n

	case "2":
		// Variable to store sorted keys
		var tmpSort []string

		// Iterate over keys to append in slice
		for k := range options {
			tmpSort = append(tmpSort, k)
		}

		// Sort slice
		sort.Strings(tmpSort)

		// Get number of metrics from user
		fmt.Printf(">> How many metrics: ")
		n, _ := reader.ReadString('\n')
		n = strings.TrimSuffix(n, "\n")

		optStr = fmt.Sprintf("%s%s,", optStr, n)

		fmt.Printf("\n")

		// Get user desired variables
		for i := 0; i < len(tmpSort); i++ {
			fmt.Printf(">> %s [y/n]: ", tmpSort[i])

			// Read user choice
			n, _ := reader.ReadString('\n')
			n = strings.TrimSuffix(n, "\n")

			// Append choice to optStr
			if strings.ToLower(n) == "y" {
				// Produce a code string like c,r,1, ... in order to know which variables user want
				optStr = fmt.Sprintf("%s%s,", optStr, options[tmpSort[i]])
			}
		}

		// Remove last "," from code str and return opt and optStr
		return opt, strings.TrimSuffix(optStr, ",")
	}

	// Return option
	return opt, ""
}

// WriteToLog - Write to log file the argument received
func WriteToLog(entry string) {
	// If the file doesn't exist, create it, or append to the file
	logFile, err := os.OpenFile("log.txt", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)

	// Make sure the logFile is closed at the end of this function
	defer logFile.Close()

	// Handle log file errors
	if err != nil {
		fmt.Printf("[Log] - Error opening log file\n")
	}

	// Write data received as argument to file
	_, err = logFile.Write([]byte(entry))

	if err != nil {
		fmt.Printf("[Log] - Error writing to log file\n")
	}
}

// GetFormatedTime - returns time in format hh:mm:ss
func GetFormatedTime() string {
	return time.Now().Format("15:04:05")
}

// GetFormatedDate - returns date in format dd/mm/aa
func GetFormatedDate() string {
	return time.Now().Format("02/01/06")
}
