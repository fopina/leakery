package main

import (
	"log"
	"os"
	"fmt"
	"flag"
	"github.com/dgraph-io/badger"
	"github.com/schollz/progressbar"
    "bufio"
    "strings"
    "time"
    //"errors"
)

const MaxInt = int(^uint(0)  >> 1) 

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
		l.Fatal(err) 
	}
}

func main() {
	logger  := MyLogger{log.New(os.Stderr, "", 0), false}

	importPtr := flag.String("import", "", "path to import directory")
	searchPtr := flag.String("search", "", "email to search for")
	debugPtr := flag.Bool("debug", false, "debug verbosity")

	flag.Parse()

	if *debugPtr {
		logger.debug = true
	}

	opts := badger.DefaultOptions
  	opts.Dir = "badger"
  	opts.ValueDir = "badger"
  	db, err := badger.Open(opts)
	logger.FatalErr(err)
	defer db.Close()

	if (*importPtr != "") {
		stat, err := os.Stat(*importPtr)
		logger.FatalErr(err)

		totalSize := int(stat.Size())
		logger.Debugln("file size", totalSize)
		
		bar := progressbar.NewOptions(
			totalSize,
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionThrottle(1 * time.Second),
			progressbar.OptionSetBytes(totalSize),
		)

		
		fd, err := os.Open(*importPtr)
	    logger.FatalErr(err)
	    defer fd.Close()

		scanner := bufio.NewScanner(fd)
		txn := db.NewTransaction(true)

		var line = ""
	    for scanner.Scan() {
	        line = scanner.Text()
	        bar.Add(len(line))
	        data := strings.Split(line, ":")
	        err = txn.Set([]byte(data[0]), []byte(data[1]))
	        if err == badger.ErrTxnTooBig {
				err = txn.Commit()
				logger.FatalErr(err)
				txn = db.NewTransaction(true)
				err = txn.Set([]byte(data[0]), []byte(data[1]))
			}
			logger.FatalErr(err)
		}

	} else if (*searchPtr != "") {
		/*
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
		*/
	} else {
		flag.Usage()
		logger.Fatal("choose an option")
	}
}