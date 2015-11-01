package main
 
import (
    "fmt"
    "net"
    "time"
		"testing"
		//"os"
    "strconv"
		"github.com/cewillis001/tftp"
)

//a six hundred char "Lorem ipsum" string

var longFile string = "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu," 
 
func checkError(err error) {
    if err  != nil {
        fmt.Println("Error: " , err)
    }
}
 
func min(a int, b int) int {
	if(a <= b) {return a}
	return b
}

func sendFullWrite(filename string, content string){
	//sends inital WRQ
	//waits for ack block 0,0 or times out
	//loop
	//  sends data
	//  waits for ack, or handles error, or resends
	//  if data sent and acknowledged was last data, end

	ServerAddr,err := net.ResolveUDPAddr("udp","127.0.0.1:10001")
	checkError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	checkError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	checkError(err)
	defer Conn.Close()	

	/* initialize variables for tracking acknowledgements and location in content*/
	index := 0
	f := []byte(content)
	max := min(512, len(f))
	data := f[index:max]
	block := []byte{0,0}

	/*send inital WRQ*/

	/*send inital WRQ*/
	tftp.SendWRQ(filename, Conn)
  buf := make([]byte, 1024)
  initialTimeout := time.After(time.Second * 15)

	InitialAck:
	for{
		select {
			case <-initialTimeout:
				fmt.Println("WRQ from ", LocalAddr, " never acknowledged")
				return
			default:
				n,_,err := Conn.ReadFromUDP(buf)
				checkError(err)
				if(err != nil) { return }
				switch buf[1]{
					case 4:
						//ACK packet recieved
						ackBlock, err := tftp.GetAckBlock(buf[0:n])
						if(err != nil){
							errCode := []byte{0,0}
							errMsg  := "badly formed ACK packet"
							tftp.SendERROR(errCode, errMsg, Conn)
							return
						}
						if(ackBlock[1] == block[1]) { break InitialAck }
					
					case 5:
						//ERROR packet recieved
						_, errMsg, err := tftp.GetError(buf[0:n])
						if(err != nil){ return }
						fmt.Println(errMsg, " recieved by ", LocalAddr)
						return			
				}
		}
	}

	SendLoop:
	for{
		//send Data
		block[1]++
		tftp.SendDATA(block, data, Conn)

		//wait for ack
		LoopAck:
		for{
			select {
				case <- time.After(time.Second * 3):
					tftp.SendDATA(block, data, Conn)
				case <- time.After(time.Second * 15):
					fmt.Println("DATA packet ", string(block), " from ", LocalAddr, " never acknowledge")
					return
				default:
					n,_,err := Conn.ReadFromUDP(buf)
					checkError(err)
					if(err != nil) { return }
					switch buf[1]{
						case 4:
							//ACK packet recieved
							ackBlock, err := tftp.GetAckBlock(buf[0:n])
							if(err != nil){
								errCode := []byte{0,0}
								errMsg  := "badly formed ACK packet"
								tftp.SendERROR(errCode, errMsg, Conn)
								return
							}
							if(ackBlock[1] == block[1]) { break LoopAck }

						case 5:
							//ERROR packet recieved
							_, errMsg, err := tftp.GetError(buf[0:n])
							if(err != nil){
								return
							}
							fmt.Println(errMsg, " recieved by ", LocalAddr)
							return
					}
			}
		}


		if(max == len(content)){
			break SendLoop
		}

		//update data
		index += 512
		max = min(index + 512, len(f))
		data = f[index : max]

	}
	
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



/* BEGIN TESTS */

func TestOneWRQ(*testing.T) {
		sendFullWrite("TestOneWRQ", "TestOneWRQ " + longFile)
} 

func TestManyWRQ(*testing.T){
	for i:=0; i < 5; i++ {
		go sendFullWrite("TestManyWRQ" + strconv.Itoa(i), "Hello, World! Just " + strconv.Itoa(i) + " world, of course")
	}
}

func TestOneRRQ(*testing.T) {
	sendFullWrite("TestOneRRQ", "Test a read!")
	sendRead("TestOneRRQ")
}

func TestManyRRQ(*testing.T) {
	sendFullWrite("TestManyRRQ", "Testing many reads!")
	for i:= 0; i < 5; i++ {
		go sendRead("TestManyRRQ")
	}
}

func TestOneOverWrite(*testing.T) {
	sendFullWrite("TestOneOverWrite", "TestOneOverWrite " + longFile)
	sendFullWrite("TestOneOverWrite", "TestOneOverWrite, overwritten " + longFile)
}
