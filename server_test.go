package main

import (
	"errors"
	"fmt"
	"github.com/cewillis001/tftp"
	"net"
	"strconv"
	"testing"
	"time"
)

//a six hundred char "Lorem ipsum" string

var longFile string = "Lorem ipsum dolor sit amet, consectetuer adipiscing elit. Aenean commodo ligula eget dolor. Aenean massa. Cum sociis natoque penatibus et magnis dis parturient montes, nascetur ridiculus mus. Donec quam felis, ultricies nec, pellentesque eu, pretium quis, sem. Nulla consequat massa quis enim. Donec pede justo, fringilla vel, aliquet nec, vulputate eget, arcu. In enim justo, rhoncus ut, imperdiet a, venenatis vitae, justo. Nullam dictum felis eu pede mollis pretium. Integer tincidunt. Cras dapibus. Vivamus elementum semper nisi. Aenean vulputate eleifend tellus. Aenean leo ligula, porttitor eu,"

func checkError(err error) {
	if err != nil {
		fmt.Println("Error: ", err)
	}
}

func min(a int, b int) int {
	if a <= b {
		return a
	}
	return b
}

func sendFullWrite(filename string, content string) {
	//sends inital WRQ
	//waits for ack block 0,0 or times out
	//loop
	//  sends data
	//  waits for ack, or handles error, or resends
	//  if data sent and acknowledged was last data, end

	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
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
	block := []byte{0, 0}

	/*send inital WRQ*/

	/*send inital WRQ*/
	tftp.SendWRQ(filename, Conn)
	packet := make(chan []byte)
	defer close(packet)
	go getPacket(Conn, packet)

	select {
	case <-time.After(15 * time.Second):
		fmt.Println("WRQ from ", LocalAddr, " never acknowledged")
		return
	case r := <-packet:
		switch r[1] {
		case 4:
			//ACK packet recieved
			ackBlock, err := tftp.GetAckBlock(r)
			if err != nil {
				errCode := []byte{0, 0}
				errMsg := "badly formed ACK packet"
				tftp.SendERROR(errCode, errMsg, Conn)
				return
			}
			if ackBlock[1] == block[1] {
				break
			}

		case 5:
			//ERROR packet recieved
			_, errMsg, err := tftp.GetError(r)
			if err != nil {
				return
			}
			fmt.Println(string(errMsg), " recieved by ", LocalAddr)
			return
		}
	}

SendLoop:
	for {
		//send Data
		block[1]++
		tftp.SendDATA(block, data, Conn)

		//wait for ack
	LoopAck:
		for {
			select {
			case <-time.After(time.Second * 3):
				tftp.SendDATA(block, data, Conn)
			case <-time.After(time.Second * 15):
				fmt.Println("DATA packet ", string(block), " from ", LocalAddr, " never acknowledged")
				return
			case r := <-packet:
				switch r[1] {
				case 4:
					//ACK packet recieved
					ackBlock, err := tftp.GetAckBlock(r)
					if err != nil {
						errCode := []byte{0, 0}
						errMsg := "badly formed ACK packet"
						tftp.SendERROR(errCode, errMsg, Conn)
						return
					}
					if ackBlock[1] == block[1] {
						break LoopAck
					}

				case 5:
					//ERROR packet recieved
					_, errMsg, err := tftp.GetError(r)
					if err != nil {
						return
					}
					fmt.Println(string(errMsg), " recieved by ", LocalAddr)
					return
				}
			}
		}

		if max == len(content) {
			break SendLoop
		}

		//update data
		index += 512
		max = min(index+512, len(f))
		data = f[index:max]

	}

}

func getPacket(Conn *net.UDPConn, out chan<- []byte) {
	for {
		buf := make([]byte, 1024)
		n, _, err := Conn.ReadFromUDP(buf)
		if err == nil {
			out <- buf[0:n]
		} else {
			return
		}
	}
}

func sendFullRead(filename string) (string, error) {
	//sends RRQ
	//loop:
	//  listens for DATA with block ahead of prev ACK packet
	//  send ACK packet for current DATA
	//  if len(current DATA) < 512, send final ACK and close
	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:10001")
	checkError(err)

	LocalAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:0")
	checkError(err)

	Conn, err := net.DialUDP("udp", LocalAddr, ServerAddr)
	checkError(err)
	defer Conn.Close()

	temp := []byte{}
	tftp.SendRRQ(filename, Conn)
	prev_block := []byte{0, 0}
	packet := make(chan []byte)
	defer close(packet)
	go getPacket(Conn, packet)

Read:
	for {
		select {
		case r := <-packet:
			switch r[1] {
			case 3:
				//DATA packet recieved
				block, data, err := tftp.GetData(r)
				if err != nil {
					return "", err
				}
				if block[1] == prev_block[1]+1 {
					tftp.SendACK(block, Conn)
					prev_block[1] = block[1]
					temp = append(temp, data...)
					if len(data) < 512 {
						break Read
					}
				} else if block[1] != 0 {
					tftp.SendACK(prev_block, Conn)
				} // else the RRQ itself failed somehow

			case 5:
				//ERROR packet recieved
				_, errMsg, err := tftp.GetError(r)
				if err != nil {
					return "", err
				}
				fmt.Println(string(errMsg), " recieved by ", LocalAddr)
				return "", errors.New(string(errMsg))
			}
		case <-time.After(1 * time.Second):
			tftp.SendACK(prev_block, Conn)

		case <-time.After(15 * time.Second):
			fmt.Println("A read request by ", LocalAddr, " timed out")
			return "", errors.New("rrq timed out")
		}
	}

	return string(temp), nil
}

/* BEGIN TESTS */

func TestOneWRQ(*testing.T) {
	sendFullWrite("TestOneWRQ", "TestOneWRQ "+longFile)
} // */

func TestManyWRQ(*testing.T) {
	for i := 0; i < 5; i++ {
		go sendFullWrite("TestManyWRQ"+strconv.Itoa(i), "Hello, World "+strconv.Itoa(i)+"! "+longFile)
	}
} // */

func TestOneRRQ(t *testing.T) {
	temp := "Test a read! " + longFile
	sendFullWrite("TestOneRRQ", temp)
	read, err := sendFullRead("TestOneRRQ")
	if err != nil {
		return
	}
	if len(temp) != len(read) {
		t.FailNow()
	} else {
		for i := 0; i < len(temp); i++ {
			if temp[i] != read[i] {
				t.FailNow()
			}
		}
	}
} // */

func TestManyRRQ(*testing.T) {
	sendFullWrite("TestManyRRQ", "Testing many reads! "+longFile)
	for i := 0; i < 5; i++ {
		go sendFullRead("TestManyRRQ")
	}
} // */

func TestOneOverWrite(*testing.T) {
	sendFullWrite("TestOneOverWrite", "TestOneOverWrite "+longFile)
	sendFullWrite("TestOneOverWrite", "TestOneOverWrite, overwritten "+longFile)
} // */
