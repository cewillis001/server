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
	"github.com/cewillis001/packet"
	//"errors"
)

func CheckError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
        os.Exit(0)
    }
}
 
/*
func ReadOpcode(header []byte) (string, error){ //accepts input, returns opcode
	if(len(header) < 2){
		return "0", errors.New("no opcode")
	}
	
	return "0", nil
} */

func main() {
	/* local memory for storing files */
	/* syntactically works: */
	m := map[string][]byte{} //creates map of empty files
	slice := []byte{1, 2, 3}
  m["file0"] = append(m["file0"],slice...)
 
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
      fmt.Println("Received ",string(buf[0:n]), " from ",addr)
 			if(buf[1] == 2){
				p, err := packet.MakeACK(0)
				CheckError(err)

				FromAddr, err := net.ResolveUDPADD("udp", "127.0.0.1:0")
				_,err = ServerConn.Write(p)
        if err != nil {
            fmt.Println("Error: ",err)
        }
			} 
    }
}
