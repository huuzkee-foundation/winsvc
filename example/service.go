// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"github.com/huuzkee-foundation/winsvc/debug"
	"github.com/huuzkee-foundation/winsvc/eventlog"
	"github.com/huuzkee-foundation/winsvc/svc"
	"fmt"
	"time"
    	"os"
    	"io"
    	"bytes"
	"os/exec"
	"log"
)

var elog debug.Log

type myservice struct{}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	fasttick := time.Tick(500 * time.Millisecond)
	slowtick := time.Tick(2 * time.Second)
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	
	const name = "QUERYHUB"
	const supports = eventlog.Error | eventlog.Warning | eventlog.Info
	err := eventlog.InstallAsEventCreate(name, supports)
	if err != nil {
		log.Printf("Install failed: %s", err)
	}
	l, err := eventlog.Open(name)
	if err != nil {
		log.Printf("Open failed: %s", err)
	}
	defer l.Close()
	err = l.Info(1, "STARTING")
	if err != nil {
		log.Printf("Info failed: %s", err)
	}
	
	cmd := exec.Command("cmd" )
	//cmd := exec.Command("echo", "'WHAT THE HECK IS UP'")



   	xr,xw  := io.Pipe()
  
 

    

	// open the out file for writing
    	outfile, err := os.Create("C:\\Users\\Marcelle\\git\\fincore\\FDM3-dev\\portal\\queryhub\\LOG_DUMP.txt")
    	if err != nil {
		err = l.Info(1, "FAILED outfile")
   	 }
   	//defer outfile.Close()
   	outfile.WriteString("HELLO WORLD\n")
   	
   	var b bytes.Buffer
	b.Write([]byte("C:\\Users\\Marcelle\\git\\fincore\\FDM3-dev\\portal\\queryhub\\activator.bat run  >> LOG.TXT"))
	fmt.Fprintf(&b, "\n")

	    
	    
   	cmd.Stdin = xr
    	cmd.Stdout = outfile
    
	err1 := cmd.Start()
	b.WriteTo( xw )
		
	if err1 != nil {
		//log.Fatal(err)
		err = l.Info(1, "FAILED")
		if err != nil {
			log.Printf("Info failed: %s", err)
		}
	}
	
	
	err = l.Info(1, "STARTED")
	if err != nil {
		log.Printf("Info failed: %s", err)
	}
	
	

	   // c1.Wait()
	    //w.Close()
    
    
    
	log.Printf("Waiting for command to finish...")
	//err = cmd.Wait()
	log.Printf("Command finished with error: %v", err)
	
	//cmd.Wait()
	
loop:
	for {
		select {
		case <-tick:
			//outfile.flush()
			beep()
			
		case c := <-r:
			switch c.Cmd {
			case svc.Interrogate:
				changes <- c.CurrentStatus
				// testing deadlock from https://code.google.com/p/winsvc/issues/detail?id=4
				time.Sleep(100 * time.Millisecond)
				changes <- c.CurrentStatus
			case svc.Stop, svc.Shutdown:
			   	outfile.WriteString("CLOSING\n")
				err = l.Info(1, "CLOSING")
				outfile.Close()
				err = l.Info(1, "CLOSED")
				break loop
			case svc.Pause:
				changes <- svc.Status{State: svc.Paused, Accepts: cmdsAccepted}
				tick = slowtick
			case svc.Continue:
				changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
				tick = fasttick
			default:
				elog.Error(1, fmt.Sprintf("unexpected control request #%d", c))
			}
		}
	}
	changes <- svc.Status{State: svc.StopPending}
	return
}

func runService(name string, isDebug bool) {
	var err error
	if isDebug {
		elog = debug.New(name)
	} else {
		elog, err = eventlog.Open(name)
		if err != nil {
			return
		}
	}
	defer elog.Close()

	elog.Info(1, fmt.Sprintf("starting %s service", name))
	run := svc.Run
	if isDebug {
		run = debug.Run
	}
	err = run(name, &myservice{})
	if err != nil {
		elog.Error(1, fmt.Sprintf("%s service failed: %v", name, err))
		return
	}
	elog.Info(1, fmt.Sprintf("%s service stopped", name))
}
