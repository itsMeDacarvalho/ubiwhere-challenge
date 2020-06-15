/*
	Author	:	Daniel Alexandre Neves de Carvalho
	Date	:	15/06/2020
	File	:	database.go
	Overview: 	Database provides some structures to handle different types
				of entries in database. It also provides functions to communicate
				with the database.
*/

package database

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/boltdb/bolt"
)

// Config type
type Config struct {
	LastAccessTime string `json:"lastAccessTime"`
}

// PerformanceOS type
type PerformanceOS struct {
	CPU      float64 `json:"cpu"`
	TotalRAM uint64  `json:"totalRAM"`
	UsedRAM  uint64  `json:"usedRAM"`
}

// Sample type
type Sample struct {
	Sample1 int `json:"sample1"`
	Sample2 int `json:"sample2"`
	Sample3 int `json:"sample3"`
	Sample4 int `json:"sample4"`
}

// SetupDB - Create a new DB with name given as argument, if DB
// dont exist already.
func SetupDB(nameDB string) (*bolt.DB, error) {
	//Try to open a new database as nameDB
	db, err := bolt.Open(nameDB+".db", 0600, nil)

	//Error handling
	if err != nil {
		fmt.Printf("[Database] - Error opening %s.db \n", nameDB)
	}

	//Create root bucket and needed buckets into root
	err = db.Update(func(tx *bolt.Tx) error {
		//Root bucket
		root, err := tx.CreateBucketIfNotExists([]byte("DB"))

		if err != nil {
			return fmt.Errorf("[Database] - Error creating root bucket: %v", err)
		}

		//OS bucket
		_, err = root.CreateBucketIfNotExists([]byte("OS"))

		if err != nil {
			return fmt.Errorf("[Database] - Error creating OS bucket into root: %v", err)
		}

		//SAMPLES bucket
		_, err = root.CreateBucketIfNotExists([]byte("SAMPLES"))

		if err != nil {
			return fmt.Errorf("[Database] - Error creating SAMPLES bucket into root: %v", err)
		}

		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("[Database] - Error performing update: %v", err)
	}

	//Retuns database created and nil since are no errors
	return db, nil
}

// SetConfig - Perform an entry on CONFIG table in database with some metadata info
func SetConfig(db *bolt.DB, config Config) error {

	//Encoding info to store as json
	configBytes, err := json.Marshal(config)

	//Handle marshall possible errors
	if err != nil {
		fmt.Printf("[DB Write] - Error encoding config data: %v", err)
	}

	//Access database and store a key/value pair as config info
	err = db.Update(func(tx *bolt.Tx) error {
		err = tx.Bucket([]byte("DB")).Put([]byte("CONFIG"), configBytes)

		//Handle possible update errors
		if err != nil {
			return fmt.Errorf("[Database] - Failed setting config: %v", err)
		}

		return nil
	})

	return err
}

// AddSystemStat - Perform a entry on OS table in database with new info
func AddSystemStat(db *bolt.DB, cpu float64, totalRAM uint64, usedRAM uint64) error {
	// Create a PerformanceOS struct in order to create a json bytes to store
	stats := PerformanceOS{CPU: cpu, TotalRAM: totalRAM, UsedRAM: usedRAM}
	statsBytes, err := json.Marshal(stats)

	// Handle json encoding errors
	if err != nil {
		return fmt.Errorf("[Database] - Error encoding stats data: %v", err)
	}

	// Update database with new CPU and RAM info
	err = db.Update(func(tx *bolt.Tx) error {
		//Write to OS table - TIME : [CPU, TotalRAM, UsedRAM]
		err := tx.Bucket([]byte("DB")).Bucket([]byte("OS")).Put([]byte(time.Now().Format("06/01/02 15:04:05")), statsBytes)

		// Handle database update error
		if err != nil {
			return fmt.Errorf("[Database] - Error inserting data to OS bucket: %v", err)
		}

		return nil
	})

	// Return err as nil if success!
	return err
}

// AddSample - Perform a entry on SAMPLE table on database with new samples
func AddSample(db *bolt.DB, samples []int) error {
	// Create a Sample struct in order to create a json bytes to store
	sample := Sample{Sample1: samples[0], Sample2: samples[1], Sample3: samples[2], Sample4: samples[3]}
	sampleBytes, err := json.Marshal(sample)

	// Handle json encoding errors
	if err != nil {
		return fmt.Errorf("[Database] - Error encoding sample data: %v", err)
	}

	// Update database with new samples from external device (sim package)
	err = db.Update(func(tx *bolt.Tx) error {
		//Write to SAMPLES table - TIME : S1 : S2 : S3 : S4
		err := tx.Bucket([]byte("DB")).Bucket([]byte("SAMPLES")).Put([]byte(time.Now().Format("06/01/02 15:04:05")), sampleBytes)

		// Handle database update error
		if err != nil {
			return fmt.Errorf("[Database] - Error inserting data to SAMPLES bucket: %v", err)
		}

		return nil

	})

	// Return err as nil if success!
	return err
}
