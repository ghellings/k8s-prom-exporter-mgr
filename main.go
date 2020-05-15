package main

import(
	"os"
	"os/signal"
	"syscall"
	"log"
	"fmt"
	"time"
	"flag"

	"github.com/ghellings/k8s-prom-exporter-mgr/app"
)

var version bool
var sleeptime int64
var configfile string
var once bool


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
  flag.Parse()
	
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
		err := exportermgr.Run()
		if err != nil {
			log.Fatal(err)
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
		time.Sleep(time.Duration(sleeptime * int64(time.Millisecond)))
	}
}