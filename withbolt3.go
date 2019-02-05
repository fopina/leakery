package main

import (
	"log"
	"os"
	"runtime"
	"fmt"
	"flag"
	bolt "github.com/coreos/bbolt"
	"github.com/schollz/progressbar/v2"
    "bufio"
    "strings"
    "time"
    //"errors"
	"encoding/binary"
	"sync"
)

type MyLogger struct {
    *log.Logger
    debug	bool
}

func (l *MyLogger) Debugln(v ...interface{}) {
	if l.debug {
		l.Output(2, fmt.Sprintln(v...)) 
	}
}

func (l *MyLogger) FatalErr(err error) {
	if err != nil {
		if l.debug {
			_, fn, line, _ := runtime.Caller(1)
			l.Fatalf("func: %s, line: %d, %v", fn, line, err) 
		} else {
			l.Fatal(err) 
		}
	}
}

var logger = MyLogger{log.New(os.Stderr, "", 0), false}

func main() {
	importPtr := flag.String("import", "", "path to import directory")
	debugPtr := flag.Bool("debug", false, "debug verbosity")

	flag.Parse()

	if *debugPtr {
		logger.debug = true
		logger.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	if (*importPtr != "") {
		stat, err := os.Stat(*importPtr)
		logger.FatalErr(err)

		logger.Debugln("file size", stat.Size())
		
		bar := progressbar.NewOptions64(
			stat.Size(),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionThrottle(1 * time.Second),
			progressbar.OptionSetBytes64(stat.Size()),
		)

		jobs := make(chan string)
		wg := new(sync.WaitGroup)
		db, err := bolt.Open("bolt.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
		logger.FatalErr(err)
		defer db.Close()

		for w := 1; w <= 3; w++ {
			wg.Add(1)
		    go addElements(jobs, wg, db, fmt.Sprintf("worker%d", w))
		}

		go func() {
			fd, err := os.Open(*importPtr)
		    logger.FatalErr(err)
		    defer fd.Close()
			scanner := bufio.NewScanner(fd)
			var line string
			for scanner.Scan() {
				line = scanner.Text()
	        	bar.Add(len(line) + 1)  // 1 or 2 for newline...?
	        	jobs <- line
	        }
	        close(jobs)
		}()

		// Now collect all the results...
	  	// But first, make sure we close the result channel when everything was processed
	  	go func() {
		    wg.Wait()
	  	}()
	} else {
		flag.Usage()
		logger.Fatal("choose an option")
	}
}

func addElements(jobs <-chan string, wg *sync.WaitGroup, db *bolt.DB, wn string) {
	defer wg.Done()

	name := []byte("bigDB")

	tx, err := db.Begin(true)
	logger.FatalErr(err)
		
	linesRead := 0
	for line := range jobs {
		linesRead += 1
	    data := strings.Split(line, ":")
		bucket, err := tx.CreateBucketIfNotExists([]byte(data[0]))
		logger.FatalErr(err)
		n, err := bucket.NextSequence()
		logger.FatalErr(err)
		err = bucket.Put(byteID(n), []byte(data[1]))
		logger.FatalErr(err)
		logger.Println(wn, data)
		// commit on every N lines
		if linesRead % 10000 == 0 {
			err = tx.Commit()
			logger.FatalErr(err)
			tx, err = db.Begin(true)
			logger.FatalErr(err)
			bucket = tx.Bucket(name)
		}
	}
}

// Create a byte slice from an uint64
func byteID(x uint64) []byte {
	b := make([]byte, 8)
	binary.BigEndian.PutUint64(b, x)
	return b
}