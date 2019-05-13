package logger

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"
)

var loger *log.Logger

var mylogpath string = ""
var currLogFile string = ""
var lock sync.Mutex

func Init(logpath string) {
	mylogpath = logpath
	if mylogpath == "" {
		return
	}
	logFile := mylogpath + "/" + time.Now().Format("2006-01-02") + ".log"
	currLogFile = logFile

	if _, errf := os.Stat(mylogpath); os.IsNotExist(errf) {
		os.MkdirAll(mylogpath, os.ModePerm) //0777也可以os.ModePerm
	}

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_RDWR, os.ModePerm)
	if err != nil {
		fmt.Println(err)
		return
	}
	loger = log.New(file, "", log.Ldate|log.Ltime|log.Lshortfile)
}

func getLoger() *log.Logger {
	if mylogpath == "" {
		return nil
	}
	lock.Lock()
	defer lock.Unlock()
	logFile := mylogpath + "/" + time.Now().Format("2006-01-02") + ".log"
	if logFile == currLogFile {
		return loger
	}
	Init(mylogpath)
	return loger
}

func Println(v ...interface{}) {
	loger = getLoger()
	if loger == nil {
		log.Print(fmt.Sprintln(v...))
		return
	}
	loger.Print(fmt.Sprintln(v...))
}

func Print(v ...interface{}) {
	loger = getLoger()
	if loger == nil {
		log.Print(fmt.Sprint(v...))
		return
	}
	loger.Print(fmt.Sprint(v...))
}

func Printf(format string, v ...interface{}) {
	loger = getLoger()
	if loger == nil {
		log.Printf(fmt.Sprintf(format, v...))
		return
	}
	loger.Printf(fmt.Sprintf(format, v...))
}
