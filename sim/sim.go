package sim

import (
	"math/rand"
	"time"
)

// GenerateSamples - This function will generate random data in order
// to simulate an external device. Lets say 4 int samples!
func GenerateSamples() []int {
	sampleData := make([]int, 4)

	//Different source for random usage
	rand.Seed(time.Now().UnixNano())

	//Collect 4 random int to generate desired sample
	for i := range sampleData {
		sampleData[i] = rand.Intn(10)
	}

	//Return sample generated previously with 4 ints
	return sampleData
}
