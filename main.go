package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/gorilla/websocket"
	"github.com/jlaffaye/ftp"
)

// ----
// The URL downloader.

type Graph struct {
	total   int
	lastTot int
}

var g Graph
var url string

// ----
// Custom io.Writer that simply tracks received bytes for graphing

func (g *Graph) Write(p []byte) (n int, err error) {
	g.total += 8 * len(p)
	return len(p), nil
}

// ----
// Download methods (FTP, HTTP, Random data)

func DownloadFTPFile(host, dir, file string) error {
	c, err := ftp.Dial(host+":21", ftp.DialWithTimeout(5*time.Second))
	if err != nil {
		return err
	}

	err = c.Login("anonymous", "anonymous")
	if err != nil {
		return err
	}

	c.ChangeDir(dir)
	res, err := c.Retr(file)
	if err != nil {
		return err
	}
	defer res.Close()

	_, err = io.Copy(&g, res)
	return err
}

func DownloadHTTPFile(url string) error {
	// Get the data
	fmt.Println("Downloading file '", url, "'")
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	_, err = io.Copy(&g, resp.Body)
	return err
}

func DownloadFile(file string) error {
	if strings.HasPrefix(file, "http://") || strings.HasPrefix(file, "https://") {
		return DownloadHTTPFile(file)
	} else if strings.HasPrefix(file, "ftp://") {
		host := file[6:]
		idx1 := strings.Index(host, "/")
		idx2 := strings.LastIndex(host, "/")
		dir := host[idx1:idx2]
		file := host[idx2+1:]
		host = host[:idx1]
		fmt.Printf("host=%s\ndir=%s\nfile=%s\n", host, dir, file)
		return DownloadFTPFile(host, dir, file)
	} else {
		fmt.Println("Unsupported URL")
	}
	return nil
}

// -----
// Randomizer

func DownloadRandomizer() {
	for {
		g.total += 8 * rand.Intn(1000000)
		time.Sleep(200 * time.Millisecond)
	}
}

// -----
// Web socket

// JSONLinkStats information
type JSONMsg struct {
	BW   uint32
	Path string
}

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func readCmd(conn *websocket.Conn) {
	for {
		// Read message from browser
		_, cmd, err := conn.ReadMessage()
		if err != nil {
			return
		}
		fmt.Printf("%s sent: %s\n", conn.RemoteAddr(), string(cmd))
		if err := DownloadFile(string(cmd)); err != nil {
			panic(err)
		}
	}
}

func printStatsOnWebSocket(w http.ResponseWriter, r *http.Request) {
	origin := r.RemoteAddr
	fmt.Println("Web Socket connection request from source=" + origin)

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("Failed to open websocket, err=%s\n", err)
		return
	}
	go readCmd(conn)

	for {
		var speed = (g.total - g.lastTot) / 1000
		g.lastTot = g.total
		msg := JSONMsg{BW: uint32(speed), Path: url}

		// Write statistics to browser
		b, err := json.Marshal(msg)
		if err = conn.WriteMessage(websocket.TextMessage, b); err != nil {
			return
		}

		fmt.Printf("Speed=%d kbps\n", speed)
		time.Sleep(1 * time.Second)
	}
}

// DisplayStats Show statistics for bandwidth
func DisplayStats() {
	http.HandleFunc("/stats", printStatsOnWebSocket)

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})
	http.HandleFunc("/app.js", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "app.js")
	})
	http.HandleFunc("/style.css", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "style.css")
	})

	http.ListenAndServe(":8080", nil)
}

// ----
// Main

func main() {
	random := flag.Bool("r", false, "Randomize bandwidth for testing")
	flag.StringVar(&url, "url", "", "Provide URL to pre-provision the UI with")
	flag.Parse()

	go DisplayStats()

	if *random {
		DownloadRandomizer()
	} else {
		for {
			time.Sleep(1 * time.Second)
		}
	}
}
