package main

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/http/cookiejar"
	"net/mail"
	"net/url"
	"os"
	"os/signal"
	"time"

	"github.com/Jeffail/gabs"
	"github.com/go-resty/resty/v2"
	"github.com/gorilla/websocket"
)

type SkipTillReader struct {
	rdr   *bufio.Reader
	delim []byte
	found bool
}

func NewSkipTillReader(reader io.Reader, delim []byte) *SkipTillReader {
	return &SkipTillReader{
		rdr:   bufio.NewReader(reader),
		delim: delim,
		found: false,
	}
}

func (str *SkipTillReader) Read(p []byte) (n int, err error) {
	if str.found {
		return str.rdr.Read(p)
	} else {
		// search byte by byte for the delimiter
	outer:
		for {
			for i := range str.delim {
				var c byte
				c, err = str.rdr.ReadByte()
				if err != nil {
					n = 0
					return
				}
				// doens't match so start over
				if str.delim[i] != c {
					continue outer
				}
			}
			str.found = true
			// we read the delimiter so add it back
			str.rdr = bufio.NewReader(io.MultiReader(bytes.NewReader(str.delim), str.rdr))
			return str.Read(p)
		}
	}
}

type ReadTillReader struct {
	rdr   *bufio.Reader
	delim []byte
	found bool
}

func NewReadTillReader(reader io.Reader, delim []byte) *ReadTillReader {
	return &ReadTillReader{
		rdr:   bufio.NewReader(reader),
		delim: delim,
		found: false,
	}
}

func (rtr *ReadTillReader) Read(p []byte) (n int, err error) {
	if rtr.found {
		return 0, io.EOF
	} else {
	outer:
		for n < len(p) {
			for i := range rtr.delim {
				var c byte
				c, err = rtr.rdr.ReadByte()
				if err != nil && n > 0 {
					err = nil
					return
				} else if err != nil {
					return
				}
				p[n] = c
				n++
				if rtr.delim[i] != c {
					continue outer
				}
			}
			rtr.found = true
			break
		}
		if n == 0 {
			err = io.EOF
		}
		return
	}
}

func printTitle() {
	fmt.Println(" ██████╗ ██╗  ██╗    ██████╗  ██████╗ ████████╗")
	fmt.Println("██╔═══██╗██║  ██║    ██╔══██╗██╔═══██╗╚══██╔══╝")
	fmt.Println("██║   ██║███████║    ██████╔╝██║   ██║   ██║   ")
	fmt.Println("██║   ██║██╔══██║    ██╔══██╗██║   ██║   ██║   ")
	fmt.Println("╚██████╔╝██║  ██║    ██████╔╝╚██████╔╝   ██║   ")
	fmt.Println(" ╚═════╝ ╚═╝  ╚═╝    ╚═════╝  ╚═════╝    ╚═╝   ")
}

func validateEmailFormat(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

type EecsohCookie struct {
	Value string `json:"value"`
}

type OhCookie struct {
	Userid  string `json:"user_id"`
	Session string `json:"session"`
}

type User struct {
	Email     string
	Signed_In bool
}

var VERSION string

func read_input(client *resty.Client) (string, string, string, int) {
	location := ""
	description := ""
	course_id := ""
	website := 0
	fmt.Printf("Enter Website Option\n1) https://oh.eecs.umich.edu/\n2) https://eecsoh.eecs.umich.edu/\n")
	fmt.Printf("Option: ")
	fmt.Scanln(&website)
	fmt.Printf("Course ID (last part of url): ")
	fmt.Scanln(&course_id)
	fmt.Printf("Location: ")
	fmt.Scanln(&location)
	fmt.Printf("Description: ")
	fmt.Scanln(&description)
	if website != 1 && website != 2 {
		return "", "", "", website
	}
	return course_id, location, description, website
}

func run_server() {
	session_cookie := ""
	user_id_cookie := ""
	office_hours_help_queue_session_cookie := ""

	http.HandleFunc("/send_session_eecsoh/", func(w http.ResponseWriter, r *http.Request) {
		session_cookie = r.PostFormValue("session")
	})

	http.HandleFunc("/send_session_oh/", func(w http.ResponseWriter, r *http.Request) {
		user_id_cookie = r.PostFormValue("user_id")
		office_hours_help_queue_session_cookie = r.PostFormValue("_office-hours-help-queue_session")
	})

	http.HandleFunc("/get_session_eecsoh/", func(w http.ResponseWriter, r *http.Request) {
		data := EecsohCookie{session_cookie}
		jData, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jData)
	})

	http.HandleFunc("/get_session_oh/", func(w http.ResponseWriter, r *http.Request) {
		data := OhCookie{user_id_cookie, office_hours_help_queue_session_cookie}
		jData, err := json.Marshal(data)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write(jData)
	})

	fmt.Printf("Starting server for login...\n")
	if err := http.ListenAndServe(":8081", nil); err != nil {
		panic(err)
	}
}

func post_queue(client *resty.Client, course_id string, location string, description string) {
	fmt.Println("Fetching login... Use the chrome extension to login")
	session := ""
	for session == "" {
		resp, err := client.R().
			SetHeaders(map[string]string{
				"accept": "*/*",
			}).
			Get("http://localhost:8081/get_session_eecsoh/")
		if err != nil {
			fmt.Println(err)
			continue
		}
		jsonParsed, err := gabs.ParseJSON(resp.Body())
		if err != nil {
			fmt.Println(err)
			continue
		}
		session = jsonParsed.Path("value").Data().(string)
		time.Sleep(500 * time.Millisecond)
	}
	client.SetCookie(&http.Cookie{
		Name:     "session",
		Value:    session,
		Path:     "/",
		Domain:   "eecsoh.eecs.umich.edu",
		MaxAge:   0,
		HttpOnly: true,
		Secure:   true,
	})
	fmt.Println("Session Fetched!")

	open := false
	enable_location_field := true
	for !open {
		fmt.Println("Checking Queue...")
		resp, err := client.R().
			SetHeaders(map[string]string{
				"accept":          "*/*",
				"accept-encoding": "gzip, deflate, br",
				"referer":         fmt.Sprintf("https://eecsoh.eecs.umich.edu/api/queues/%s", course_id),
				"user-agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36",
			}).
			Get(fmt.Sprintf("https://eecsoh.eecs.umich.edu/api/queues/%s", course_id))
		if err != nil {
			fmt.Println(err)
			continue
		}
		jsonParsed, err := gabs.ParseJSON(resp.Body())
		if err != nil {
			fmt.Println(err)
			continue
		}
		open = jsonParsed.Path("open").Data().(bool)
		enable_location_field = jsonParsed.Path("config.enable_location_field").Data().(bool)
		if !open {
			fmt.Println("Queue Closed. Retrying in 500ms")
			time.Sleep(500 * time.Millisecond)
		}
	}
	fmt.Println("Queue Open!")
	done := false
	for !done {
		body := map[string]string{
			"description": description,
			"location":    "(disabled)",
		}
		if enable_location_field {
			body = map[string]string{
				"description": description,
				"location":    location,
			}
		}

		resp, err := client.R().
			SetHeaders(map[string]string{
				"origin":          "https://eecsoh.eecs.umich.edu",
				"user-agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36",
				"accept":          "*/*",
				"accept-encoding": "gzip, deflate, br",
				"referer":         fmt.Sprintf("https://eecsoh.eecs.umich.edu/api/queues/%s", course_id),
			}).
			SetBody(body).
			Post(fmt.Sprintf("https://eecsoh.eecs.umich.edu/api/queues/%s/entries", course_id))
		if err != nil {
			fmt.Println(err)
			continue
		}
		fmt.Println(resp.String())
		jsonParsed, err := gabs.ParseJSON(resp.Body())
		if err != nil {
			fmt.Println(err)
			continue
		}
		_, ok := jsonParsed.Path("open").Data().(bool)
		if ok {
			done = true
		} else {
			fmt.Println("Queue Entry Failed. Retrying in 200ms")
		}
		time.Sleep(200 * time.Millisecond)
	}
	fmt.Println("Queue Successfully Entered!")
}

func post_queue_oh(client *resty.Client, course_id string, location string, description string) {
	fmt.Println("Fetching login... Use the chrome extension to login")
	user_id_cookie := ""
	office_hours_help_queue_session_cookie := ""
	for user_id_cookie == "" || office_hours_help_queue_session_cookie == "" {
		resp, err := client.R().
			SetHeaders(map[string]string{
				"accept": "*/*",
			}).
			Get("http://localhost:8081/get_session_oh/")
		if err != nil {
			continue
		}
		// fmt.Println(resp.String())
		jsonParsed, err := gabs.ParseJSON(resp.Body())
		if err != nil {
			continue
		}
		user_id_cookie = jsonParsed.Path("user_id").Data().(string)
		office_hours_help_queue_session_cookie = jsonParsed.Path("session").Data().(string)
		time.Sleep(500 * time.Millisecond)
	}
	fmt.Println("Session Fetched!")

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)
	u := "wss://oh.eecs.umich.edu/cable"
	fmt.Printf("connecting to %s\n", u)
	jar, _ := cookiejar.New(nil)
	cookie1 := http.Cookie{
		Name:     "user_id",
		Value:    user_id_cookie,
		Path:     "/",
		Domain:   "oh.eecs.umich.edu",
		MaxAge:   0,
		HttpOnly: false,
		Secure:   false,
	}
	cookie2 := http.Cookie{
		Name:     "_office-hours-help-queue_session",
		Value:    office_hours_help_queue_session_cookie,
		Path:     "/",
		Domain:   "oh.eecs.umich.edu",
		MaxAge:   0,
		HttpOnly: false,
		Secure:   false,
	}
	u1, err := url.Parse("https://oh.eecs.umich.edu/")
	jar.SetCookies(u1, []*http.Cookie{&cookie1, &cookie2})

	header := http.Header(map[string][]string{
		"Host":                     {"oh.eecs.umich.edu"},
		"Origin":                   {"https://oh.eecs.umich.edu"},
		"Cache-Control":            {"no-cache"},
		"Sec-WebSocket-Extensions": {"permessage-deflate; client_max_window_bits"},
		"Sec-WebSocket-Protocol":   {"actioncable-v1-json, actioncable-unsupported"},
		"User-Agent":               {"Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.82 Safari/537.36"},
	})
	dialer := websocket.DefaultDialer
	dialer.Jar = jar
	c, resp, err := dialer.Dial(u, header)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	fmt.Println(resp)

	done := make(chan struct{})

	go func() {
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf("recv: %s\n", message)
		}
	}()

	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()

	stringJSON := fmt.Sprintf(`{"command": "subscribe",	"identifier": "{\"channel\":\"QueueChannel\",\"id\":%s}"}`, course_id)
	err = c.WriteMessage(websocket.TextMessage, []byte(stringJSON))
	if err != nil {
		log.Println("Error during writing to websocket:", err)
		return
	}

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			stringJSON := fmt.Sprintf(`{"command":"message","identifier":"{\"channel\":\"QueueChannel\",\"id\":%s}","data":"{\"location\":\"%s\",\"description\":\"%s\",\"action\":\"new_request\"}"}`, course_id, location, description)
			err = c.WriteMessage(websocket.TextMessage, []byte(stringJSON))
			if err != nil {
				log.Println("write:", err)
				return
			}
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return
		}
	}
}

func main() {
	client := resty.New()
	client.SetTimeout(4 * time.Second)
	printTitle()
	fmt.Printf("\n\n")

	course_id, location, description, website := read_input(client)

	go run_server()

	if website == 1 {
		post_queue_oh(client, course_id, location, description)
	} else if website == 2 {
		post_queue(client, course_id, location, description)
	} else {
		return
	}
}
