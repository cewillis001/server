/*
	A simple TFTP server for Igneous by CE Willis, 28 OCt 2015
	-stores files only in memory
	-only implements octet mode
	-source of very simple udp client, not TFTP complient:
		https://varshneyabhi.wordpress.com/2014/12/23/simple-udp-clientserver-in-golang/
*/

package main

import (
	"fmt"
	"net"
	"os"
	"github.com/cewillis001/tftp"
)

func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
}

func DeleteChannel(done <-chan *net.UDPAddr, ch map[string]chan []byte){
	for r := range done{
		fmt.Println(r, " is done")
		//_,present := ch[r.String()]
		//if(present) { fmt.Println(r, " channel in map ") }
		close(ch[r.String()])
		delete(ch, r.String())
	}
}

func WriteFile(in <-chan *tftp.File, m map[string][]byte){
  for r := range in {
    _,present := m[r.Name]
    if(!present) {
      m[r.Name] = r.Data
    }else{
      errCode := []byte {0, 6}
      errMsg  := "No overwriting allowed"
      tftp.SendERRORTo(errCode, errMsg, r.Conn, r.Raddr)
    }
  }
}

func main() {
	/* local memory for storing files */
	m := map[string][]byte{} //creates map of empty files

	/* map of channels to communicate with routines handling read/write requests */
	ch := map[string](chan []byte){}
	/* channel for the one deleter go routine */
	done := make(chan *net.UDPAddr)
	defer close(done)
	go DeleteChannel(done, ch)

	/* channel for the one writer go routine */
 	newFile := make(chan *tftp.File)
	defer close(newFile)
	go WriteFile(newFile, m) 

	/* Lets prepare a address at any address at port 10001*/   
	ServerAddr,err := net.ResolveUDPAddr("udp",":10001")
	CheckError(err)
 
	/* Now listen at selected port */
	ServerConn, err := net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	defer ServerConn.Close()

	buf := make([]byte, 1024)
 
	for {
		n,addr,err := ServerConn.ReadFromUDP(buf)
		CheckError(err)
		//fmt.Println("Received ",string(buf[0:n]), " from ",addr)

		if(n < 2 || buf[0] != 0 || buf[1] == 0 || buf[1] > 5){
			tftp.SendERRORTo([]byte{0, 0}, "bad opcode", ServerConn, addr)
			continue
		}

		switch buf[1]{
			case 1:
				//handle RRQ
				fmt.Println("Recived ", string(buf[0:n]), " from ", addr, " as RRQ")
				filename, err := tftp.GetRRQname(buf[0:n])
				if(err != nil){
					//badly formed rrq packet
					errCode := []byte {0, 0}
					errMsg  := "Badly formed RRQ packet"
					tftp.SendERRORTo(errCode, errMsg, ServerConn, addr)
				} else {
					_,present := m[filename]
					if(present){
						ch[addr.String()] = make(chan []byte) 
						go tftp.HandleRRQ(filename, ch[addr.String()], done, ServerConn, addr, m[filename])
					} else {
						errCode := []byte {0, 1}
						errMsg  := filename + " not found"
						tftp.SendERRORTo(errCode, errMsg, ServerConn, addr)
					}
				}

			case 2:
				//handle WRQ
				fmt.Println("Recieved ", string(buf[0:n]), " from ", addr, " as WRQ")
				filename, err := tftp.GetWRQname(buf[0:n])
				if(err != nil){
					errCode := []byte {0, 0}
					errMsg  := "Badly formed WRQ packet"
					tftp.SendERRORTo(errCode, errMsg, ServerConn, addr)
				} else {
					_,present := m[filename]
					if(present){
						errCode := []byte {0, 6}
						errMsg  := "No overwriting allowed"
						tftp.SendERRORTo(errCode, errMsg, ServerConn, addr)
					}else{
						ch[addr.String()] = make(chan []byte)
						//fmt.Println(addr)
						_,present := ch[addr.String()]
						if(present){
							fmt.Println(addr, " added successfully to map")
						}
						go tftp.HandleWRQ(filename, ch[addr.String()], done, ServerConn, addr, newFile)
					}	
				}

			case 3:
				//handle DATA
				fmt.Println("Recieved ", string(buf[0:n]), " from ", addr, " as DATA")
				_,present := ch[addr.String()]
				if(present){
					//fmt.Println("channel is present")
					ch[addr.String()] <- buf[0:n]
				} else {
					errCode := []byte {0, 5}
					errMsg  := "No write request associated with this return address"
					fmt.Println(errMsg, " ", addr)
					tftp.SendERRORTo(errCode, errMsg, ServerConn, addr)
				} 

			case 4:
				//handle ACK
				fmt.Println("Recieved ", string(buf[0:n]), " from ", addr, " as ACK")
				_,present := ch[addr.String()]
				if(present){
					ch[addr.String()] <- buf[0:n]
				} else {
					errCode := []byte {0, 5}
					errMsg  := "No read request associated with this return address"
					tftp.SendERRORTo(errCode, errMsg, ServerConn, addr)
				}

			case 5:
				//handle ERROR
				/* 
					all errors should term connection, EXCEPT if the error is because
					"the source port of a recieved tftp is incorrect. in this case
					an errr tftp is sent to the originating host"
				*/
				fmt.Println("Recieved ", string(buf[0:n]), " from ", addr, " as ERROR")
				_,present := ch[addr.String()]
				if(present){
					close(ch[addr.String()])
					delete(ch, addr.String())
				}
		}
	}
}
