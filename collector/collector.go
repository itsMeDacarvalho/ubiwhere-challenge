package collector

import (
	"fmt"
	"os"
	"time"
	"ubiwhere-challenge/sim"
)

/*CollectSamples : get samples from external simulator
and store them in a local db. Local DB is a file.
*/
func CollectSamples() {
	//Create neu time var
	t := time.Now()
	hour, minute, second := t.Clock()

	//Create new file to store db collected info if doesnt exist
	db, _ := os.Create("db_collectedData.txt")

	//Make sure file closes at the end of this function
	defer db.Close()

	//Create and format info to database... first time
	dataToStore := fmt.Sprintf("%d:%d:%d", hour, minute, second)

	//After we get random samples and format them in order to store
	for _, v := range sim.GenerateSamples() {
		dataToStore = fmt.Sprintf("%s\t%d", dataToStore, v)
	}

	//We write the file (database) with info collected
	db.WriteString(dataToStore + "\n")

	fmt.Println("DB write with success!")
}
