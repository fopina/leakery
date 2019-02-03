package main

import (
	"log"
	"os"
	"runtime"
	"fmt"
	"flag"
	"github.com/schollz/progressbar/v2"
    "bufio"
    "strings"
    "time"
    //"errors"
    "database/sql"
    _ "github.com/mattn/go-sqlite3"
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

    db, err := sql.Open("sqlite3", "./sqlite.db")
	logger.FatalErr(err)
	defer db.Close()
	_, err = db.Query("CREATE TABLE IF NOT EXISTS leaks (EMAIL VARCHAR(255), PASSWORD VARCHAR(255));")
	logger.FatalErr(err)
	_, err = db.Query("PRAGMA journal_mode=OFF;")
	logger.FatalErr(err)
	_, err = db.Query("PRAGMA synchronous=OFF;")
	logger.FatalErr(err)
	
	// _, err = db.Query("CREATE INDEX email_index on leaks (email);")
	// logger.FatalErr(err)
	// _, err = db.Query("DROP INDEX email_index on leaks;")
	// logger.FatalErr(err)
	tx, err := db.Begin()
	logger.FatalErr(err)

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

		
		fd, err := os.Open(*importPtr)
	    logger.FatalErr(err)
	    defer fd.Close()

		scanner := bufio.NewScanner(fd)

		line := ""
		linesRead := 0

	    for scanner.Scan() {
	        line = scanner.Text()
	        bar.Add(len(line) + 1)  // 1 or 2 for newline...?
	        data := strings.Split(line, ":")
	        _, err = db.Query("INSERT INTO leaks VALUES ( ?, ? )", data[0], data[1])
			linesRead += 1

	        if linesRead % 10000 == 0 {
	        	tx.Commit()
	        	tx, err = db.Begin()
				logger.FatalErr(err)
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