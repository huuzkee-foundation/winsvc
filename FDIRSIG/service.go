// Copyright 2012 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// +build windows

package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/huuzkee-foundation/winsvc/debug"
	"github.com/huuzkee-foundation/winsvc/eventlog"
	"github.com/huuzkee-foundation/winsvc/svc"
	"io"
	"os"
	"strconv"
	"strings"
	"time"
	//"bytes"
	"crypto/sha256"
	"encoding/base64"
	"golang.org/x/exp/utf8string"
	"log"
	"os/exec"
	"path/filepath"

	//"math"
)

var elog debug.Log

type myservice struct{}

type FileDesc struct {

	//RecId		int64
	RootId        string
	FileType      string
	FileName      string
	FileExtension string
	FileDir       string
	FilePath      string
	FileSize      string
	ModTimeLocal  string
}

//////////////////////////////////////////////////////////

type FileNameHash struct {
	RecId        string
	FileName     string
	FileNameHash string
}

//---------------------------------------------------------
type FileDirHash struct {
	RecId       string
	FilDir      string
	FileDirHash string
}

//---------------------------------------------------------
type FilePathHash struct {
	RecId        string
	FilePath     string
	FilePathHash string
}

//---------------------------------------------------------
type FileHash struct {
	RecId    string
	FileHash string
}

//---------------------------------------------------------
type FileSize struct {
	RecId    string
	FileSize int64
}

//---------------------------------------------------------
type FileStatus struct {
	RecId      string
	FileStatus string
}

//---------------------------------------------------------

/////////////////////////////////////////////////////////////////////////////////////////////

func procdirs(outfile *os.File) {

	ORG_ROOT := "sanfrancisco.huzzlee.com//"

	const filechunk = 8192 // we settle for 8KB

	outfile.WriteString("\r\nSTART PROCFILES \r\n\r\n")

	db, err := sql.Open("mysql", "fdiscovery:F1tPar0la@/fdiscovery")
	if err != nil {
		outfile.WriteString(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.dirhash ///
	stDirHash, err := db.Prepare("INSERT INTO fdiscovery.dirhash  VALUES( ?, ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stDirHash.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.rdirhash ///
	stRDirHash, err := db.Prepare("INSERT INTO fdiscovery.rdirhash  VALUES( ?, ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stRDirHash.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.odirhash ///
	stODirHash, err := db.Prepare("INSERT INTO fdiscovery.odirhash  VALUES( ?, ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stODirHash.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.dirstatus ///
	stDirStatus, err := db.Prepare("INSERT INTO fdiscovery.dirstatus  VALUES( ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stDirStatus.Close() //

	filequery, err := db.Prepare("SELECT RecId, RootId, FileType, FileName, FileExtension,	FileDir, FilePath, FileSize, ModTimeLocal FROM fdiscovery.sourcefiles WHERE RecId = ?")
	if err != nil {
		log.Fatal(err)
	}
	var (
		RecId         int64
		RootId        string
		FileType      string
		FileName      string
		FileExtension string
		Dir           string
		FilePath      string
		FileSize      string
		ModTimeLocal  string
		//FileNameHash	string
		//FileDirHash	string
		//FilePathHash	string
		//FileHash	string
		//FileStatus	string
		//FileSizeNum	int64
	)

	FileStatus := "NEW"

	for i := 7430; i < 1438884; i++ {
		err = filequery.QueryRow(i).Scan(&RecId, &RootId, &FileType, &FileName, &FileExtension, &Dir, &FilePath, &FileSize, &ModTimeLocal)
		if err != nil {
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
		}

		if FileType == "DIR" {

			outfile.WriteString("\r\nFILENO: ")
			outfile.WriteString(strconv.Itoa(int(i)))
			//outfile.WriteString( "\r\nFILENAME: " )
			//outfile.WriteString( FileName )
			outfile.WriteString("\r\nFILEPATH: ")
			outfile.WriteString(FilePath)

			///===============================================================================
			DirVec := []byte(Dir)
			DirHasher := sha256.New()
			DirHasher.Write(DirVec)
			DirHash := base64.URLEncoding.EncodeToString(DirHasher.Sum(nil))

			DirU8 := utf8string.NewString(Dir)
			RDir := DirU8.Slice(3, DirU8.RuneCount())
			RDirVec := []byte(RDir)
			RDirHasher := sha256.New()
			RDirHasher.Write(RDirVec)
			RDirHash := base64.URLEncoding.EncodeToString(RDirHasher.Sum(nil))

			//outfile.WriteString( "\r\nRDir: " )
			//outfile.WriteString( RDir )

			ODir := ORG_ROOT + Dir
			ODirVec := []byte(ODir)
			ODirHasher := sha256.New()
			ODirHasher.Write(ODirVec)
			ODirHash := base64.URLEncoding.EncodeToString(ODirHasher.Sum(nil))

			//outfile.WriteString( "\r\nODir: " )
			//outfile.WriteString( ODir )

			_, err = stDirHash.Exec(RecId, DirHash, Dir) // Insert data from fd
			if err != nil {
				outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
			}
			_, err = stRDirHash.Exec(RecId, RDirHash, RDir) // Insert data from fd
			if err != nil {
				outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
			}
			_, err = stODirHash.Exec(RecId, ODirHash, ODir) // Insert data from fd
			if err != nil {
				outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
			}
			///===============================================================================

			_, err = stDirStatus.Exec(RecId, FileStatus) // Insert data from fd
			if err != nil {
				outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
			}
			///===============================================================================

		}

		///===============================================================================

	}

}

////////////////////////////////////////////////////////////////////////////////////
////////////////////////////////////////////////////////////////////////////////////

func procfiles(outfile *os.File) {

	ORG_ROOT := "sanfrancisco.huzzlee.com//"

	const filechunk = 8192 // we settle for 8KB

	outfile.WriteString("\r\nSTART PROCFILES \r\n\r\n")

	db, err := sql.Open("mysql", "fdiscovery:F1tPar0la@/fdiscovery")
	if err != nil {
		outfile.WriteString(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.filenamehash ///
	stFileNameHash, err := db.Prepare("INSERT INTO fdiscovery.filenamehash  VALUES( ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stFileNameHash.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.filedirhash ///
	stFileDirHash, err := db.Prepare("INSERT INTO fdiscovery.filedirhash  VALUES( ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stFileDirHash.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.filerdirhash ///
	stFileRDirHash, err := db.Prepare("INSERT INTO fdiscovery.filerdirhash  VALUES( ?, ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stFileRDirHash.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.fileodirhash ///
	stFileODirHash, err := db.Prepare("INSERT INTO fdiscovery.fileodirhash  VALUES( ?, ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stFileODirHash.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.filepathhash ///
	stFilePathHash, err := db.Prepare("INSERT INTO fdiscovery.filepathhash  VALUES( ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stFilePathHash.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.filerpathhash ///
	stFileRPathHash, err := db.Prepare("INSERT INTO fdiscovery.filerpathhash  VALUES( ?, ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stFileRPathHash.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.fileopathhash ///
	stFileOPathHash, err := db.Prepare("INSERT INTO fdiscovery.fileopathhash  VALUES( ?, ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stFileOPathHash.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.filehash ///
	stFileHash, err := db.Prepare("INSERT INTO fdiscovery.filehash  VALUES( ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stFileHash.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.filesize ///
	stFileSize, err := db.Prepare("INSERT INTO fdiscovery.filesize  VALUES( ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stFileSize.Close() //
	//----------------------------------------------------------------------
	// Prepare statement for inserting data into fdiscovery.filestatus ///
	stFileStatus, err := db.Prepare("INSERT INTO fdiscovery.filestatus  VALUES( ?, ? )") //
	if err != nil {
		outfile.WriteString(err.Error()) //
	}
	defer stFileStatus.Close() //

	filequery, err := db.Prepare("SELECT RecId, RootId, FileType, FileName, FileExtension,	FileDir, FilePath, FileSize, ModTimeLocal FROM fdiscovery.sourcefiles WHERE RecId = ?")
	if err != nil {
		log.Fatal(err)
	}
	var (
		RecId         int64
		RootId        string
		FileType      string
		FileName      string
		FileExtension string
		FileDir       string
		FilePath      string
		FileSize      string
		ModTimeLocal  string
		//FileNameHash	string
		//FileDirHash	string
		//FilePathHash	string
		//FileHash	string
		//FileStatus	string
		//FileSizeNum	int64
	)

	FileStatus := "NEW"

	for i := 4; i < 1438884; i++ {
		err = filequery.QueryRow(i).Scan(&RecId, &RootId, &FileType, &FileName, &FileExtension, &FileDir, &FilePath, &FileSize, &ModTimeLocal)
		if err != nil {
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
		}
		//outfile.WriteString( "\r\nFILENO: " )
		//outfile.WriteString( strconv.Itoa( int( i) ) );
		//outfile.WriteString( "\r\nFILENAME: " )
		//outfile.WriteString( FileName )
		//outfile.WriteString( "\r\nFILEPATH: " )
		//outfile.WriteString( FilePath )

		///===============================================================================
		FileNameVec := []byte(FileName)
		FileNameHasher := sha256.New()
		FileNameHasher.Write(FileNameVec)
		FileNameHash := base64.URLEncoding.EncodeToString(FileNameHasher.Sum(nil))
		_, err = stFileNameHash.Exec(RecId, FileNameHash) // Insert data from fd
		if err != nil {
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
		}
		///===============================================================================
		FileDirVec := []byte(FileDir)
		FileDirHasher := sha256.New()
		FileDirHasher.Write(FileDirVec)
		FileDirHash := base64.URLEncoding.EncodeToString(FileDirHasher.Sum(nil))

		FileDirU8 := utf8string.NewString(FileDir)
		FileRDir := FileDirU8.Slice(3, FileDirU8.RuneCount())
		FileRDirVec := []byte(FileRDir)
		FileRDirHasher := sha256.New()
		FileRDirHasher.Write(FileRDirVec)
		FileRDirHash := base64.URLEncoding.EncodeToString(FileRDirHasher.Sum(nil))

		//outfile.WriteString( "\r\nFileRDir: " )
		//outfile.WriteString( FileRDir )

		FileODir := ORG_ROOT + FileDir
		FileODirVec := []byte(FileODir)
		FileODirHasher := sha256.New()
		FileODirHasher.Write(FileODirVec)
		FileODirHash := base64.URLEncoding.EncodeToString(FileODirHasher.Sum(nil))

		//outfile.WriteString( "\r\nFileODir: " )
		//outfile.WriteString( FileODir )

		_, err = stFileDirHash.Exec(RecId, FileDirHash) // Insert data from fd
		if err != nil {
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
		}
		_, err = stFileRDirHash.Exec(RecId, FileRDirHash, FileRDir) // Insert data from fd
		if err != nil {
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
		}
		_, err = stFileODirHash.Exec(RecId, FileODirHash, FileODir) // Insert data from fd
		if err != nil {
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
		}
		///===============================================================================

		FilePathVec := []byte(FilePath)
		FilePathHasher := sha256.New()
		FilePathHasher.Write(FilePathVec)
		FilePathHash := base64.URLEncoding.EncodeToString(FilePathHasher.Sum(nil))

		FilePathU8 := utf8string.NewString(FilePath)
		FileRPath := FilePathU8.Slice(3, FilePathU8.RuneCount())
		FileRPathVec := []byte(FileRPath)
		FileRPathHasher := sha256.New()
		FileRPathHasher.Write(FileRPathVec)
		FileRPathHash := base64.URLEncoding.EncodeToString(FileRPathHasher.Sum(nil))

		//outfile.WriteString( "\r\nFileRPath: " )
		//outfile.WriteString( FileRPath )

		FileOPath := ORG_ROOT + FilePath
		FileOPathVec := []byte(FileOPath)
		FileOPathHasher := sha256.New()
		FileOPathHasher.Write(FileOPathVec)
		FileOPathHash := base64.URLEncoding.EncodeToString(FileOPathHasher.Sum(nil))

		//outfile.WriteString( "\r\nFileOPath: " )
		//outfile.WriteString( FileOPath )

		_, err = stFilePathHash.Exec(RecId, FilePathHash) // Insert data from fd
		if err != nil {
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
		}
		_, err = stFileRPathHash.Exec(RecId, FileRPathHash, FileRPath) // Insert data from fd
		if err != nil {
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
		}
		_, err = stFileOPathHash.Exec(RecId, FileOPathHash, FileOPath) // Insert data from fd
		if err != nil {
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
		}
		///===============================================================================
		FileSizeNum, err := strconv.ParseInt(FileSize, 10, 64)
		//FileSizeNum, err := strconv.Atoi( FileSize )
		_, err = stFileSize.Exec(RecId, FileSizeNum) // Insert data from fd
		if err != nil {
			outfile.WriteString("\r\nFileSize: ")
			outfile.WriteString(FileSize)
			outfile.WriteString("\r\n")
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
		}
		///===============================================================================
		_, err = stFileStatus.Exec(RecId, FileStatus) // Insert data from fd
		if err != nil {
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
		}
		///===============================================================================
		FileHash := "DIR"
		RECY := "D:\\$RECYCLE.BIN\\"

		if strings.Contains(FilePath, RECY) {
			outfile.WriteString("\r\n PROBLEM: ")
			outfile.WriteString(FilePath)
		}

		/*
		       if FileType == "FILE" {
		       if true != strings.Contains(FilePath, RECY)  {
		   	    outfile.WriteString( "\r\nFILE: " )
		   	    outfile.WriteString( FilePath )
		   	    File, err1 := os.Open( FilePath )
		   		if err1 != nil {
		   	    		outfile.WriteString( "\r\nFILE ERROR !!!!\r\n" )
		   			outfile.WriteString(err1.Error()) // proper error handling instead of panic in your app
		   	    		outfile.WriteString( "\r\n" )
		   		}
		   	    defer File.Close()
		   	    FileHasher := sha256.New()
		   	    // calculate the file size
		   	    info, _ := File.Stat()
		   	    ifilesize := info.Size()
		   	    blocks := uint64(math.Ceil(float64(ifilesize) / float64(filechunk)))
		   	    blocksize := int(math.Min(filechunk, float64(ifilesize-int64(i*filechunk))))
		   	    buf := make([] byte, blocksize)
		   	    for i := uint64(0); i < blocks; i++ {
		   		  File.Read(buf)
		   		  FileHasher.Write( buf )
		   	    }
		   	    FileHash = base64.URLEncoding.EncodeToString(FileHasher.Sum(nil))
		       }
		       }

		*/
		_, err = stFileHash.Exec(RecId, FileHash) // Insert data from fd
		if err != nil {
			outfile.WriteString("\r\nDBERROR\r\n")
			outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
			outfile.WriteString("\r\n")
		}

		///===============================================================================

	}

}

//////////////////////////////////////////////////////////

func walkpath(path string, f os.FileInfo, err error, outfile *os.File, fchan chan FileDesc) error {

	filetype := "FILE"
	if f.IsDir() {
		filetype = "DIR"
	}
	//outfile.WriteString( filetype );
	//outfile.WriteString( ",\t\"")
	//outfile.WriteString( f.Name() );
	//outfile.WriteString( "\",\t")
	//outfile.WriteString( filepath.Ext(path) );
	//outfile.WriteString( ",\t\"")
	//outfile.WriteString( filepath.Dir(path) );
	//outfile.WriteString( "\",\t\"")
	//outfile.WriteString( f.ModTime().String() );
	//outfile.WriteString( "\",\t")
	//outfile.WriteString( strconv.Itoa( int(f.Size()) ) );
	//outfile.WriteString( ",\t\"")
	//outfile.WriteString( path );
	//outfile.WriteString( "\"\r\n")

	fdesc := FileDesc{"0000", filetype, f.Name(), filepath.Ext(path), filepath.Dir(path), path, strconv.Itoa(int(f.Size())), f.ModTime().String()}
	fchan <- fdesc

	return nil
}

func filework(outfile *os.File, rootpath string, fchan chan FileDesc) {

	outfile.WriteString("\r\nSTART FILEWALKER \r\n")

	walker := func(path string, f os.FileInfo, err error) error {
		walkpath(path, f, err, outfile, fchan)
		return nil
	}
	filepath.Walk(rootpath, walker)

	outfile.WriteString("\r\nSTOP FILEWALKER \r\n")

}

func dbwork(outfile *os.File, fchan chan FileDesc) {

	outfile.WriteString("\r\nSTART DBCONN \r\n\r\n")

	db, err := sql.Open("mysql", "fdiscovery:F1tPar0la@/fdiscovery")
	if err != nil {
		outfile.WriteString(err.Error()) // Just for example purpose. You should use proper error handling instead of panic
	}
	defer db.Close()

	//_, err = db.Exec("DELETE FROM squarenum")

	// Prepare statement for inserting data
	//stmtIns, err := db.Prepare("INSERT INTO squareNum VALUES( ?, ? )") // ? = placeholder
	stmtIns, err := db.Prepare("INSERT INTO SourceFiles VALUES( ?, ?, ?, ?, ?, ?, ?, ?, ? )") // ? = placeholder
	if err != nil {
		outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
	}
	defer stmtIns.Close() // Close the statement when we leave main() / the program terminates

	// Prepare statement for reading data
	//stmtOut, err := db.Prepare("SELECT squareNumber FROM squarenum WHERE number = ?")
	//if err != nil {
	//	outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
	//}
	//defer stmtOut.Close()

	// Insert square numbers for 0-24 in the database
	//for i := 0; i < 25; i++ {
	//    _, err = stmtIns.Exec(i, (i * i)) // Insert tuples (i, i^2)
	//    if err != nil {
	//		outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
	//    }
	//}

	recid := 1

	for {
		select {
		case fd := <-fchan:
			recid = recid + 1
			_, err = stmtIns.Exec(strconv.Itoa(int(recid)), fd.RootId, fd.FileType, fd.FileName, fd.FileExtension, fd.FileDir, fd.FilePath, fd.FileSize, fd.ModTimeLocal) // Insert data from fd
			if err != nil {
				outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
			}
			//default:
			//      outfile.WriteString("NO FILE INFOR RECEIVED\r\n")
		}
	}

	//var squareNum int // we "scan" the result in here

	// Query the square-number of 7
	//err = stmtOut.QueryRow(7).Scan(&squareNum) // WHERE number = 13
	//if err != nil {
	//	outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
	//}
	//outfile.WriteString("The square number of 7 is: ")
	//outfile.WriteString( strconv.Itoa(squareNum))
	//outfile.WriteString("\r\n")

	// Query another number.. 1 maybe?
	//err = stmtOut.QueryRow(4).Scan(&squareNum) // WHERE number = 1
	//if err != nil {
	//outfile.WriteString(err.Error()) // proper error handling instead of panic in your app
	//}
	// outfile.WriteString("The square number of 4 is: ")
	//outfile.WriteString(strconv.Itoa(squareNum))
	//outfile.WriteString("\r\n")

	outfile.WriteString("\r\nSTOP DBCONN \r\n")

	return

}

func launcher(cmdstid *io.PipeReader, outfile *os.File, winlog *eventlog.Log, service string) {

	outfile.WriteString("\r\nSTARTING CMD LAUNCHER for ")
	outfile.WriteString(service)
	outfile.WriteString(" \r\n  \r\n")

	cmdName := "cmd"
	cmdArgs := []string{" "}
	cmd := exec.Command(cmdName, cmdArgs...)
	cmd.Stdin = cmdstid
	cmd.Stdout = outfile
	errCmd := cmd.Start()
	if errCmd != nil {
		errWl := winlog.Info(1, "cmd.Start() FAILED")
		if errWl != nil {
			log.Printf("Info failed: %s", errWl)
		}
	}
	errWl := winlog.Info(1, "CMD LAUNCHER STARTED")
	if errWl != nil {
		log.Printf("Info failed: %s", errWl)
	}
	cmd.Wait()
}

func (m *myservice) Execute(args []string, r <-chan svc.ChangeRequest, changes chan<- svc.Status) (ssec bool, errno uint32) {
	const cmdsAccepted = svc.AcceptStop | svc.AcceptShutdown | svc.AcceptPauseAndContinue
	changes <- svc.Status{State: svc.StartPending}
	fasttick := time.Tick(500 * time.Millisecond)
	slowtick := time.Tick(2 * time.Second)
	tick := fasttick
	changes <- svc.Status{State: svc.Running, Accepts: cmdsAccepted}

	QUERYHUB_SERVICE := "DIRSIG-SERVICE"
	//QUERYHUB_ROOT 			:= "C:\\play\\portal\\queryhub\\"
	QUERYHUB_ROOT := "D:\\"
	QUERYHUB_LOG := QUERYHUB_ROOT + QUERYHUB_SERVICE + ".LOG"
	//QUERYHUB_ACTIVATOR		:= QUERYHUB_ROOT + "activator.bat"
	QUERYHUB_SEMAPHORE_RUNNING := QUERYHUB_ROOT + QUERYHUB_SERVICE + ".IS.RUNNING"
	QUERYHUB_SEMAPHORE_STOPPED := QUERYHUB_ROOT + QUERYHUB_SERVICE + ".IS.STOPPED"

	//filechannel := make(chan FileDesc)

	const supports = eventlog.Error | eventlog.Warning | eventlog.Info
	err := eventlog.InstallAsEventCreate(QUERYHUB_SERVICE, supports)
	if err != nil {
		log.Printf("Event Log Install failed: %s", err)
	}
	winlog, err := eventlog.Open(QUERYHUB_SERVICE)
	if err != nil {
		log.Printf("Event Log Open failed: %s", err)
	}
	defer winlog.Close()
	err = winlog.Info(1, "STARTING")
	if err != nil {
		log.Printf("Event Log Info failed: %s", err)
	}

	// open the out file for writing
	cmdstdout, err := os.Create(QUERYHUB_LOG)
	if err != nil {
		err = winlog.Info(1, "FAILED outfile")
		if err != nil {
			log.Printf("Event Log Info failed: %s", err)
		}
	}
	defer cmdstdout.Close()

	//_,cmdcon  := io.Pipe()
	//cmdstdin,cmdcon  := io.Pipe()

	//go launcher( cmdstdin, 	cmdstdout, winlog, QUERYHUB_SERVICE )
	//go dbwork( cmdstdout )

	cmdstdout.WriteString("\r\n")
	cmdstdout.WriteString(QUERYHUB_SERVICE + " is Running ")
	cmdstdout.WriteString("\r\n")

	//cmdcon.Write( []byte("\r\n") )
	//cmdcon.Write( []byte("cd ") )
	//cmdcon.Write( []byte( QUERYHUB_ROOT ) )
	//cmdcon.Write( []byte("\r\n") )
	//cmdcon.Write( []byte("del ") )
	//cmdcon.Write( []byte( QUERYHUB_SEMAPHORE_STOPPED ) )
	//cmdcon.Write( []byte("\r\n") )

	semaphore, err := os.Create(QUERYHUB_SEMAPHORE_RUNNING)
	semaphore.WriteString("\r\n")
	semaphore.WriteString(QUERYHUB_SERVICE + " is Running ")
	semaphore.WriteString("\r\n")

	//go filework( cmdstdout, QUERYHUB_ROOT, filechannel )
	//go dbwork( cmdstdout, filechannel )
	//go procfiles(cmdstdout)
	go procdirs(cmdstdout)

	//cmdcon.Write( []byte( QUERYHUB_ACTIVATOR ) )
	//cmdcon.Write( []byte(" \"run -Dhttp.port=80  -Dhttps.port=443\"") )
	//cmdcon.Write( []byte("\r\n") )

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
				semaphore.Close()
				os.Rename(QUERYHUB_SEMAPHORE_RUNNING, QUERYHUB_SEMAPHORE_STOPPED)
				err = winlog.Info(1, "SERVICE CLOSED")
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
