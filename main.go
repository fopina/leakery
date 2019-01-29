package main

import (
	"log"
	"os"
	"fmt"
	"flag"
	"github.com/xyproto/simplebolt"
	"github.com/schollz/progressbar"
    "bufio"
    "strings"
    "time"
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

	db, err := simplebolt.New("bolt.db")
	logger.FatalErr(err)
	defer db.Close()

	kv, err := simplebolt.NewKeyValue(db, "bigDB")
	logger.FatalErr(err)

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

		var line = ""
	    for scanner.Scan() {
	        line = scanner.Text()
	        bar.Add(len(line))
	        data := strings.Split(line, ":")
	        err = kv.Set(data[0], data[1])
	        logger.FatalErr(err)
	    }

	} else if (*searchPtr != "") {
		r, err := kv.Get(*searchPtr)
		if err != nil {
			logger.Fatal(err)
		}
		logger.Println(r)
	} else {
		flag.Usage()
		logger.Fatal("choose an option")
	}
}