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
    "errors"
	"encoding/binary"
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

func main() {
	logger  := MyLogger{log.New(os.Stderr, "", 0), false}

	importPtr := flag.String("import", "", "path to import directory")
	searchPtr := flag.String("search", "", "email to search for")
	debugPtr := flag.Bool("debug", false, "debug verbosity")
	resumePtr := flag.Bool("resume", false, "resume from last checkpoint")

	flag.Parse()

	if *debugPtr {
		logger.debug = true
		logger.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	db, err := bolt.Open("bolt.db", 0600, &bolt.Options{Timeout: 1 * time.Second})
	logger.FatalErr(err)
	defer db.Close()

	if (*importPtr != "") {
		stat, err := os.Stat(*importPtr)
		logger.FatalErr(err)

		logger.Debugln("file size", stat.Size())
		
		fd, err := os.Open(*importPtr)
	    logger.FatalErr(err)
	    defer fd.Close()

		scanner := bufio.NewScanner(fd)

		name := []byte("bigDB")

		tx, err := db.Begin(true)
		logger.FatalErr(err)
		
		line := ""
		linesRead := uint(0)
		lastRead := 0
		// this wouldn't be required if currentNum was exposed by ProgressBar
		// could also use reflect to access it but sounds dirtier...
		currentNum := int64(0)
		bytesRead := func(data []byte, atEOF bool) (advance int, token []byte, err error) {
			// dirty hack to get the exact bytes read
			a, b, c := bufio.ScanLines(data, atEOF)
			lastRead = a
			return a, b, c
		}
		scanner.Split(bytesRead)
		if (*resumePtr) {
			sbucket, err := tx.CreateBucketIfNotExists([]byte("==status=="))
			logger.FatalErr(err)
			r := sbucket.Get([]byte("x"))
			if len(r) > 0 {
				currentNum, _ = binary.Varint(r)
				logger.Println("Searching checkpoint", currentNum)
				fd.Seek(currentNum, 0)
			}
		}

		bar := progressbar.NewOptions64(
			stat.Size() - currentNum,
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionThrottle(1 * time.Second),
			progressbar.OptionSetBytes64(stat.Size() - currentNum),
		)

	    for scanner.Scan() {
	        line = scanner.Text()
	        linesRead += 1
	        currentNum += int64(lastRead)
	        bar.Add(lastRead)
	        data := strings.Split(line, ":")
	        bucket, err := tx.CreateBucketIfNotExists([]byte(data[0]))
			logger.FatalErr(err)
			n, err := bucket.NextSequence()
			logger.FatalErr(err)
			err = bucket.Put(byteID(int64(n)), []byte(data[1]))
			logger.FatalErr(err)
			// commit on every N lines
			if linesRead % 10000 == 0 {
				sbucket, err := tx.CreateBucketIfNotExists([]byte("==status=="))
				logger.FatalErr(err)
				err = sbucket.Put([]byte("x"), byteID(currentNum))
				logger.FatalErr(err)
				err = tx.Commit()
				logger.FatalErr(err)
				tx, err = db.Begin(true)
				logger.FatalErr(err)
				bucket = tx.Bucket(name)
			}
		}

	} else if (*searchPtr != "") {
		name := []byte("bigDB")
		err = db.Update(func(tx *bolt.Tx) error {
			if _, err := tx.CreateBucketIfNotExists(name); err != nil {
				return errors.New("Could not create bucket: " + err.Error())
			}
			return nil
		})
		var val string
		err = db.View(func(tx *bolt.Tx) error {
			bucket := tx.Bucket([]byte("bigDB"))
			if bucket == nil {
				return errors.New("bucket not found")
			}
			byteval := bucket.Get([]byte(*searchPtr))
			if byteval == nil {
				return errors.New("key not found")
			}
			val = string(byteval)
			return nil // Return from View function
		})
		logger.FatalErr(err)
		logger.Println(val)
	} else {
		flag.Usage()
		logger.Fatal("choose an option")
	}
}

// Create a byte slice from an uint64
func byteID(x int64) []byte {
	b := make([]byte, 8)
	binary.PutVarint(b, x)
	return b
}
