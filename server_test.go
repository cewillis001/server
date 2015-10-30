package main
 
import (
    "fmt"
    "net"
    //"time"
		"testing"
		//"os"
    //"strconv"
		"github.com/cewillis001/tftp"
)
 
func checkError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
    }
}
 
func sendWrite(filename string, content []byte, conn *net.UDPConn, addr *net.UDPAddr){
	
}


func TestOneWRQ(*testing.T) {
	shortFile := "Hello world"
	shortFilename := "hw"

	ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1:10001")
	checkError(err)
 
	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	checkError(err)
 
	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	checkError(err)
 
	defer Conn.Close()
	tftp.SendWRQ(shortFilename, Conn)
	buf :=make([]byte, 1024)
	_,_,err = Conn.ReadFromUDP(buf)
	checkError(err)
	if(err != nil) {return }
	block, err := tftp.GetAckBlock(buf)
	checkError(err)
	if(err != nil) {return}
	if(len(block) >=2 && block[1] != 0){
		fmt.Println("wrong block number in ACK to WRQ")
	}
	tftp.SendDATA([]byte{0,1}, []byte(shortFile), Conn)

}
