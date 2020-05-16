package main

import(
	"os"
	"os/signal"
	"syscall"
	"fmt"
	"time"
	"flag"

	log "github.com/sirupsen/logrus"
	"github.com/ghellings/k8s-prom-exporter-mgr/app"
)

var version bool
var sleeptime int64
var configfile string
var once bool
var loglevel string


const (
	versioninfo = "v0.0.1"
)

type loopinterface interface{
	Run() error
}

func main() {
	flag.BoolVar(&version, "version", false, "k8s-prom-exporter-mgr version")
	flag.Int64Var(&sleeptime, "sleeptime", 1000, "Sleep time in loop")
	flag.StringVar(&configfile, "configfile", "./k8s-prom-exporter-mgr.conf", "Full path to configfile")
	flag.BoolVar(&once, "once", false, "Run once")
	flag.StringVar(&loglevel, "loglevel", "info", "The level of log output (trace,debug,info,warn,error)")
  flag.Parse()
  
  // Log as JSON instead of the default ASCII formatter.
  log.SetFormatter(&log.JSONFormatter{})

  // Output to stdout instead of the default stderr
  // Can be any io.Writer, see below for File example
  log.SetOutput(os.Stdout)

  // Only log the Info severity or above.
  switch loglevel {
  case "trace": 
  	log.SetLevel(log.TraceLevel)
  case "debug":
  	log.SetLevel(log.DebugLevel)
  case "info":
  	log.SetLevel(log.InfoLevel)
  case "warn":
  	log.SetLevel(log.WarnLevel)
  case "error":
  	log.SetLevel(log.ErrorLevel)
  } 
	if version {
		fmt.Printf("k8s-prom-exporter-mgr version %s\n", versioninfo)
		return
	}
	config, err := exportermgr.ReadConfig(configfile)
	if err != nil {
		log.Fatal(err)
	}
	exportermgr := exportermgr.New(config)
	switch {
	case once:
		log.Info("Run once and exit")
		err := exportermgr.Run()
		if err != nil {
			log.Error(err)
		}
		return
	default:
		loop(exportermgr)
	}
}

func loop(loop loopinterface) {
	c := make(chan os.Signal, 1)
	r := make(chan bool)

	signal.Notify(c, syscall.SIGHUP)

	go func(){
		for sig := range c {
			fmt.Println(sig)
			r <- true
		}
	}()
	for {
		select {
		case msg := <-r:
			log.Printf("Recieved HUP. Reloading: %#v\n", msg)
				return
			default:
		}
		err := loop.Run()
		if err != nil {
			log.Println(err)
			return
		}
		log.Debugf("Sleeping for %d seconds",sleeptime)
		time.Sleep(time.Duration(sleeptime * int64(time.Millisecond)))
	}
}