package main

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
)

// HTTPContent struct to hold HTTP request details
type HTTPContent struct {
	headers map[string]interface{}
	method  string
	path    string
	version string
	body    string
}

func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	fmt.Println("Logs from your program will appear here!")

	// Uncomment this block to pass the first stage
	listener, err := net.Listen("tcp", "0.0.0.0:4221")
	if err != nil {
		fmt.Println("Failed to bind to port 4221:", err)
		os.Exit(1)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	httpContent := HTTPContent{}
	buf := make([]byte, 1024)
	_, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading from connection:", err)
		return
	}

	content := strings.Split(strings.TrimSpace(string(buf)), "\r\n")
	body := strings.Split(strings.TrimSpace(string(buf)), "\r\n\r\n")[1]
	path := strings.Split(content[0], " ")
	httpContent.method = path[0]
	httpContent.path = path[1]
	httpContent.version = path[2]
	httpContent.body = body

	tempMap := make(map[string]interface{}, 10)
	for x := 1; x < len(content); x++ {
		headers := strings.Split(strings.TrimSpace(content[x]), ": ")
		if len(headers) > 1 {
			tempMap[headers[0]] = headers[1]
		}
	}

	httpContent.headers = tempMap

	paths := strings.Split(httpContent.path, "/")[1]
	switch httpContent.method {
	case "GET":
		switch paths {
		case "", "/":
			conn.Write([]byte("HTTP/1.1 200 OK\r\n\r\n"))
		case "/echo", "echo":
			lastPath := strings.Split(httpContent.path, "/echo")
			lastPath1 := strings.TrimLeft(lastPath[len(lastPath)-1], "/")
			lastPathLen := len(lastPath1)
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", lastPathLen, lastPath1)
		case "/user-agent", "user-agent":
			userAgent := fmt.Sprintf("%v", httpContent.headers["User-Agent"])
			userAgentLen := len(userAgent)
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: text/plain\r\nContent-Length: %d\r\n\r\n%s", userAgentLen, userAgent)
		case "/files", "files":
			args := os.Args[2]
			file := strings.Split(httpContent.path, "/")[len(strings.Split(httpContent.path, "/"))-1]
			path := filepath.Join(args, file)
			fileByte, err := os.ReadFile(path)
			if err != nil {
				conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
				return
			}
			fmt.Fprintf(conn, "HTTP/1.1 200 OK\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(fileByte), fileByte)
		default:
			conn.Write([]byte("HTTP/1.1 404 Not Found\r\n\r\n"))
		}
	case "POST":
		switch paths {
		case "/files", "files":
			args := os.Args[2]
			file := strings.Split(httpContent.path, "/")[len(strings.Split(httpContent.path, "/"))-1]
			path := filepath.Join(args, file)
			err := os.WriteFile(path, []byte(strings.Trim(httpContent.body, "\x00")), os.ModeAppend)
			if err == nil {
				fmt.Fprintf(conn, "HTTP/1.1 201 Created\r\nContent-Type: application/octet-stream\r\nContent-Length: %d\r\n\r\n%s", len(strings.Trim(httpContent.body, "\x00")), strings.Trim(httpContent.body, "\x00"))
			}
		}
	default:
		fmt.Println("Method not supported")
	}
}
