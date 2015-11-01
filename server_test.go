package main
 
import (
    "fmt"
    "net"
    //"time"
		"testing"
		//"os"
    "strconv"
		"github.com/cewillis001/tftp"
)
 
func checkError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
    }
}
 
func sendWrite(filename string, content string){

	//This sends one short file, expects one acknowlegement from server

  ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1:10001")
  checkError(err)

  LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
  checkError(err)

  Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
  checkError(err)

  defer Conn.Close()
  tftp.SendWRQ(filename, Conn)
  buf := make([]byte, 1024)
  n,_,err := Conn.ReadFromUDP(buf)
  checkError(err)
  if(err != nil) {return }
  block, err := tftp.GetAckBlock(buf[0:n])
  checkError(err)
  if(err != nil) {return}
  if(block[1] != 0){
    fmt.Println("wrong block number in ACK to WRQ")
  }
  tftp.SendDATA([]byte{0,1}, []byte(content), Conn)
}

func sendRead(filename string){
	//this expects one short  (single packet) file
  ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1:10001")
  checkError(err)

  LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
  checkError(err)

  Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
  checkError(err)

  defer Conn.Close()
  tftp.SendRRQ(filename, Conn)

  buf := make([]byte, 1024)
  _,_,err = Conn.ReadFromUDP(buf)
  checkError(err)
  if(err != nil) {return }

  block, data, err := tftp.GetData(buf)
  checkError(err)
  if(err != nil) {return}
  if(len(block) >=2 && block[1] != 0){
    fmt.Println("wrong block number in ACK to WRQ")
  }
	fmt.Println("Read ", string(data))
	tftp.SendACK(block, Conn)
}

func TestOneWRQ(*testing.T) {
		sendWrite("TestOneWRQ", "Hello, World! Just one world, of course")
}

func TestManyWRQ(*testing.T){
	for i:=0; i < 5; i++ {
		go sendWrite("TestManyWRQ" + strconv.Itoa(i), "Hello, World! Just " + strconv.Itoa(i) + " world, of course")
	}
}

func TestOneRRQ(*testing.T) {
	sendWrite("TestOneRRQ", "Test a read!")
	sendRead("TestOneRRQ")
}
