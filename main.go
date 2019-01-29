package main

import (
	"log"
	"os"
	"runtime"
	"fmt"
	"flag"
	"github.com/globalsign/mgo"
	"github.com/fopina/progressbar"
    "bufio"
    "strings"
    "time"
    //"errors"
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

type LoginData struct {
	Username	string
	Password	string
}

func main() {
	logger  := MyLogger{log.New(os.Stderr, "", 0), false}

	importPtr := flag.String("import", "", "path to import directory")
	searchPtr := flag.String("search", "", "email to search for")
	debugPtr := flag.Bool("debug", false, "debug verbosity")

	flag.Parse()

	if *debugPtr {
		logger.debug = true
		logger.SetFlags(log.LstdFlags | log.Lshortfile)
	}

	session, err := mgo.Dial("localhost:28000")
	logger.FatalErr(err)
	defer session.Close()

	if (*importPtr != "") {
		stat, err := os.Stat(*importPtr)
		logger.FatalErr(err)

		logger.Debugln("file size", stat.Size())
		
		bar := progressbar.NewOptions(
			stat.Size(),
			progressbar.OptionSetRenderBlankState(true),
			progressbar.OptionThrottle(1 * time.Second),
			progressbar.OptionSetBytes(stat.Size()),
		)

		
		fd, err := os.Open(*importPtr)
	    logger.FatalErr(err)
	    defer fd.Close()

		scanner := bufio.NewScanner(fd)

		c := session.DB("leaks").C("emails")

		index := Index{
		    Key: []string{"username"},
		    Unique: true,
		    DropDups: true,
		    Background: true, // See notes.
		    Sparse: true,
		}
		err = c.EnsureIndex(index)
		logger.FatalErr(err)

		line := ""
		linesRead := 0
		bulk := c.Bulk()

	    for scanner.Scan() {
	        line = scanner.Text()
	        bar.Add(len(line) + 1)  // 1 or 2 for newline...?
	        data := strings.Split(line, ":")
	        bulk.Insert(LoginData{Username: data[0], Password: data[1]})
			linesRead += 1

	        if linesRead % 10000 == 0 {
	        	bulk.Run()
	        }
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