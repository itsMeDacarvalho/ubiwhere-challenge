package toolset

import (
	"bufio"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strconv"
	"strings"
	"time"
	"ubiwhere-challenge/collector"
	"ubiwhere-challenge/database"
	"ubiwhere-challenge/sim"

	"github.com/boltdb/bolt"
)

// PrintAllSampleData show in terminal well formated SAMPLE table with all data
func PrintAllSampleData(db *bolt.DB) error {
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

// PrintAllOSData show in terminal well formated OS table with all data
func PrintAllOSData(db *bolt.DB) error {
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

// StoreDataOS get info from OS and perform an entry on database
func StoreDataOS(db *bolt.DB) error {
	infoRAM := collector.GetRAM()
	err := database.AddSystemStat(db, collector.GetCPU(), infoRAM[0], infoRAM[1])

	return err
}

// StoreSample get samples from simulator and perform an entry on database
func StoreSample(db *bolt.DB) error {
	err := database.AddSample(db, sim.GenerateSamples())

	return err
}

// GetLastN  take a database to lookup and a code string with variables and number
// of metrics to return
func GetLastN(db *bolt.DB, codeStr string) (map[int][]int, error) {
	// Map with type of sample as key to array of last n values
	sampleData := make(map[int][]int)

	// Code string to array - this code maps which variables user want to see
	codeArray := strings.Split(codeStr, ",")

	// Desired numberof metrics. This come from first element of codeStr
	desiredN, _ := strconv.Atoi(codeArray[0])

	// Temporary structs to decode json
	var tmpSamples database.Sample
	var tmpOS database.PerformanceOS

	// Start database transaction
	err := db.View(func(tx *bolt.Tx) error {
		// Get SAMPLES table data
		sampleTable := tx.Bucket([]byte("DB")).Bucket([]byte("SAMPLES"))
		osTable := tx.Bucket([]byte("DB")).Bucket([]byte("OS"))

		//Create new cursor for SAMPLES and OS table and point to last entry
		sampleCursor := sampleTable.Cursor()
		osCursor := osTable.Cursor()

		// Point to last entry in SAMPLES and OS tables
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

			// Create final map with desired info
			if Find(codeArray, "c") {
				sampleData[desiredN-i] = append(sampleData[desiredN-i], int(math.Round(tmpOS.CPU)))
			}
			if Find(codeArray, "r") {
				sampleData[desiredN-i] = append(sampleData[desiredN-i], int(tmpOS.UsedRAM))
			}
			if Find(codeArray, "1") {
				sampleData[desiredN-i] = append(sampleData[desiredN-i], tmpSamples.Sample1)
			}
			if Find(codeArray, "2") {
				sampleData[desiredN-i] = append(sampleData[desiredN-i], tmpSamples.Sample2)
			}
			if Find(codeArray, "3") {
				sampleData[desiredN-i] = append(sampleData[desiredN-i], tmpSamples.Sample3)
			}
			if Find(codeArray, "4") {
				sampleData[desiredN-i] = append(sampleData[desiredN-i], tmpSamples.Sample4)
			}

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
func PrintLastN(data map[int][]int, codeStr string) {
	iterationsLoop := len(data[1])

	// Display a nice header for our table
	FormatTableHeader(codeStr, iterationsLoop)

	// Sort keys to show correct info
	var keys []int
	for k := range data {
		keys = append(keys, k)
	}
	sort.Ints(keys)

	// Start printing info in table
	for i := 1; i <= len(data); i++ {
		fmt.Printf("|%d\t", i)
		for j := 0; j < iterationsLoop; j++ {
			fmt.Printf("| %d \t\t", data[i][j])
		}
		fmt.Printf("|\n")
	}

	// Display bottom of table in a nice way too
	fmt.Printf("+-------")

	for i := 0; i < iterationsLoop; i++ {
		fmt.Printf("----------------")
	}
	fmt.Printf("+\n")
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
		return opt, fmt.Sprintf("%s,c,r,1,2,3,4", n)

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

// FormatTableHeader displays a custom table header for desired info
func FormatTableHeader(headerInfo string, numberRows int) {
	decodeStr := map[string]string{
		"c": "CPU (%)",
		"r": "RAM (Mb)",
		"1": "Sample 1",
		"2": "Sample 2",
		"3": "Sample 3",
		"4": "Sample 4",
	}

	// Convert header info string into an array in order to iterate it
	codeArray := strings.Split(headerInfo, ",")

	fmt.Printf("+-------")

	for i := 0; i < numberRows; i++ {
		fmt.Printf("----------------")
	}
	fmt.Printf("+\n")

	// Get variable name from decode string map and produce the table header
	fmt.Printf("|# \t")
	for i := 1; i < len(codeArray); i++ {
		fmt.Printf("| %s \t", decodeStr[codeArray[i]])
	}

	fmt.Printf("|\n")
	fmt.Printf("+-------")

	for i := 0; i < numberRows; i++ {
		fmt.Printf("----------------")
	}
	fmt.Printf("+\n")
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

// Find takes a slice and looks for an element in it. If found it will
// return it's key, otherwise it will return -1 and a bool of false.
func Find(slice []string, val string) bool {
	for _, item := range slice {
		if item == val {
			return true
		}
	}
	return false
}
