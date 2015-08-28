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
    	//"strings"
    	"io"
    	"bytes"
	"os/exec"
	"log"
)

var elog debug.Log

type myservice struct{}

	
func launcher(cmdstid *io.PipeReader,  outfile *os.File, from string) {

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
	err = l.Info(1, from )
	if err != nil {
		log.Printf("Info failed: %s", err)
	}
	
	//app := "cmd"
        //arg00 := "C:\tmp\test.bat"
        //arg01 := " "
        
	//app := "C:\\Program Files (x86)\\Java\\jdk17055\\bin\\java.exe"
        //arg00 := "C:\\tmp\\test.bat"
        //arg01 := " "
        
        cmdName := "cmd"
       // cmdName := "git"
	cmdArgs := []string{"C:\\tmp\\test.bat", " ", " "}
	//cmdArgs := []string{"rev-parse", "--verify", "HEAD"}
	
	cmd := exec.Command( cmdName, cmdArgs... )
	cmd.Stdin = cmdstid
	
	//cmd.Stdin = strings.NewReader( arg00 )
	//cmd := exec.Command("echo", "'WHAT THE HECK IS UP'")


      
	// open the out file for writing
    	//outfile, err = os.Create("C:\\Go\\usr\\marcelle\\src\\src\\github.com\\huuzkee-foundation\\winsvc\\example\\LOG_DUMP.txt")
    	//if err != nil {
	//	err = l.Info(1, "FAILED outfile")
   	// }
   	//defer outfile.Close()
   	outfile.WriteString( from )
   	
   	//var b bytes.Buffer
	//b.Write([]byte("C:\\Users\\Marcelle\\git\\fincore\\FDM3-dev\\portal\\queryhub\\activator.bat run  >> LOG.TXT"))
	//fmt.Fprintf(&b, "\n")
	//b.WriteTo( xw )
	
   	//cmd.Stdin = xr
    	cmd.Stdout = outfile
    
	err1 := cmd.Start()
	cmd.Wait()
	//err1 := cmd.Run()
	//b.WriteTo( xw )
		
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
	   // w.Close()
    
    
    
	log.Printf("Waiting for command to finish...")
	//err = cmd.Wait()
	log.Printf("Command finished with error: %v", err)
	

	

}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	fasttick := time.Tick(500 * time.Millisecond)
	slowtick := time.Tick(2 * time.Second)
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}
	
	
	// open the out file for writing
	outfile, err := os.Create("C:\\Go\\usr\\marcelle\\src\\src\\github.com\\huuzkee-foundation\\winsvc\\example\\LOG_DUMP.txt")

	if err != nil {
		//err = l.Info(1, "FAILED outfile")
	}
   	defer outfile.Close()
   	
   	
	cmdstdin,cmdcon  := io.Pipe()
	go launcher( cmdstdin, 	outfile, "STARTING LAUNCHER\r\n")
	
	testfile, err := os.Create( "C:\\tmp\\test.txt.SEMAPHORE" )
	
	var b bytes.Buffer
	//b.Write([]byte("C:\\Users\\Marcelle\\git\\fincore\\FDM3-dev\\portal\\queryhub\\activator.bat run  >> LOG.TXT"))
	b.Write([]byte("C:\\tmp\\test.bat"))
	//fmt.Fprintf(&b, "\n")
	
	outfile.WriteString( "\r\n" )
   	outfile.WriteString( b.String() )	
	outfile.WriteString( "\r\n" )
	
	testfile.WriteString( "\r\n" )
	testfile.WriteString( b.String() )	
	testfile.WriteString( "\r\n" )
   	
   	cmdcon.Write( []byte("\r\n") )
   	cmdcon.Write( []byte(b.String()) )
   	cmdcon.Write( []byte("\r\n") )
   	
	//b.WriteTo( cmdcon )
	//b.WriteTo( cmdcon )	
	//b.WriteTo( cmdcon )
	
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
				testfile.Close()
			        os.Rename( "C:\\tmp\\test.txt.SEMAPHORE", "C:\\tmp\\test.txt")
			   	//outfile.WriteString("CLOSING\n")
				//err = l.Info(1, "CLOSING")
				//outfile.Close()
				//err = l.Info(1, "CLOSED")
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
