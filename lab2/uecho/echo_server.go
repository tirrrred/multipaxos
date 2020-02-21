// +build !solution

// Leave an empty line above this comment.
package uecho

import (
	"fmt"
	"log"
	"net"
	"strings"
	"unicode"
)

// UDPServer implements the UDP server specification found at
// https://github.com/dat520-2020/assignments/blob/master/lab2/README.md#udp-server
type UDPServer struct {
	conn *net.UDPConn
	// TODO(student): Add fields if needed
}

// NewUDPServer returns a new UDPServer listening on addr. It should return an
// error if there was any problem resolving or listening on the provided addr.
func NewUDPServer(addr string) (*UDPServer, error) {
	// TODO(student): Implement
	//NewUDPServer("localhost:12110") returns a UDPServer struct
	laddr, err := net.ResolveUDPAddr("udp", addr)
	fmt.Printf("echo_server - NewUDPServer: laddr = %v, %T\n", laddr, laddr)
	if err != nil {
		log.Fatal("Address given failed", err)
		return nil, err
	}
	udpServerConn, err := net.ListenUDP("udp", laddr)
	fmt.Printf("echo_server - NewUDPServer: udpServerConn = %v, %T\n", udpServerConn, udpServerConn)
	if err != nil {
		log.Fatal("Listing for UDP on socket failed:", err)
		return nil, err
	}
	return &UDPServer{udpServerConn}, nil
}

// ServeUDP starts the UDP server's read loop. The server should read from its
// listening socket and handle incoming client requests as according to the
// the specification.
func (u *UDPServer) ServeUDP() {
	// TODO(student): Implement
	serverConn := u.conn
	fmt.Printf("echo_server - ServeUDP: serverConn = %v, %T\n", serverConn, serverConn)

	defer serverConn.Close() //look into defer

	// run loop forever (or until ctrl-c)
	for {
		buffer := make([]byte, 512, 1024)
		n, srcadr, err := u.conn.ReadFromUDP(buffer[0:])
		if err != nil {
			log.Fatal("Reading from UDP connection failed", err)
			fmt.Println("Connection from", srcadr)
			//return something?
		}
		for _, v := range buffer[0:n] {
			fmt.Print(string(v))
		}
		clientInputstr := string(buffer[0:n])
		clientInput := strings.Split(clientInputstr, "|:|")

		/*cmd := clientInput[0]
		text := clientInput[1]
		fmt.Println("\nCommand:", cmd)
		fmt.Println("Text:", text)*/

		output, err := serverReply(clientInput)

		if err != nil {
			log.Fatal("Something went wrong", err)
		}

		//b := []byte(output)
		u.conn.WriteToUDP([]byte(output), srcadr)
	}

}

func serverReply(clientInput []string) (string, error) {
	var reply string
	var err error

	cmd := clientInput[0]
	text := clientInput[1]

	fmt.Println("\nCommand:", cmd)
	fmt.Println("Text:", text)

	switch {
	case len(clientInput) > 2:
		reply = "Unknown command"
	case text == "":
		reply = ""
	case cmd == "UPPER":
		reply = strings.ToUpper(text)
	case cmd == "LOWER":
		reply = strings.ToLower(text)
	case cmd == "CAMEL":
		reply = camelCase(text)
	case cmd == "ROT13":
		reply = rot13(text)
	case cmd == "SWAP":
		reply = swap(text)
	default:
		reply = "Unknown command"
	}

	fmt.Println(reply)
	fmt.Println("\n----------------------")

	return reply, err
}

func camelCase(text string) string {
	lowString := strings.ToLower(text)
	b := []byte(lowString)

	b[0] = byte(unicode.ToUpper(rune(b[0])))

	for i, v := range b {
		if unicode.IsSpace(rune(v)) {
			b[i+1] = byte(unicode.ToUpper(rune(b[i+1])))
		} else {
			continue
		}
	}
	s := string(b)
	return s
}

func swap(text string) string {
	b := []byte(text)
	for i, v := range b {
		if unicode.IsSpace(rune(v)) {
			continue
		}
		if unicode.IsUpper(rune(v)) {
			b[i] = byte(unicode.ToLower(rune(b[i])))
		}
		if unicode.IsLower(rune(v)) {
			b[i] = byte(unicode.ToUpper(rune(b[i])))
		}
	}
	s := string(b)
	return s
}

func rot13(text string) string {
	b := []byte(text)
	for i := 0; i < len(b); i++ {
		if (b[i] >= 'A' && b[i] < 'N') || (b[i] >= 'a' && b[i] < 'n') {
			b[i] += 13
		} else if (b[i] >= 'N' && b[i] <= 'Z') || (b[i] >= 'n' && b[i] <= 'z') {
			b[i] -= 13
		}
	}
	s := string(b)
	return s
}

// socketIsClosed is a helper method to check if a listening socket has been
// closed.
func socketIsClosed(err error) bool {
	if strings.Contains(err.Error(), "use of closed network connection") {
		return true
	}
	return false
}
