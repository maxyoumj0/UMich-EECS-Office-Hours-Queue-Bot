package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"strings"
	"time"

	"github.com/go-resty/resty/v2"
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

func read_input() (string, string, string, string, string) {
	username := ""
	password := ""
	location := ""
	description := ""
	url := ""
	fmt.Println("Enter UMich Login")
	fmt.Printf("Username: ")
	fmt.Scanln(&username)
	fmt.Printf("Password: ")
	fmt.Scanln(&password)
	fmt.Printf("OH URL: ")
	fmt.Scanln(&url)
	fmt.Printf("Location: ")
	fmt.Scanln(&location)
	fmt.Printf("Description: ")
	fmt.Scanln(&description)
	return username, password, url, location, description
}

func main() {
	username, password, url, location, description := read_input()
	client := resty.New()
	client.SetTimeout(4 * time.Second)

	// resp, err := client.R().
	// 	SetHeaders(map[string]string{
	// 		"Host":       "oh.eecs.umich.edu",
	// 		"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36",
	// 		"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
	// 	}).
	// 	Get("https://oh.eecs.umich.edu/course_queues/786")
	// client.SetCookies(resp.Cookies())

	resp, err := client.R().
		SetHeaders(map[string]string{
			"Host":       "weblogin.umich.edu",
			"User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36",
			"Accept":     "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.9",
		}).
		Get("https://weblogin.umich.edu/?cosign-shibboleth.umich.edu&https://shibboleth.umich.edu/idp/Authn/RemoteUser?conversation=e1s1")
	if err != nil {
		panic(err)
	}
	if resp.StatusCode() != 200 {
		fmt.Println("cosgin get response not 200")
		return
	}
	client.SetCookies(resp.Cookies())

	resp, err = client.R().
		SetHeaders(map[string]string{
			"Host":         "weblogin.umich.edu",
			"Referer":      "https://weblogin.umich.edu/?cosign-shibboleth.umich.edu&https://shibboleth.umich.edu/idp/Authn/RemoteUser?conversation=e1s1",
			"User-Agent":   "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36",
			"Content-Type": "application/x-www-form-urlencoded",
		}).
		SetFormData(map[string]string{
			"ref":      "https://shibboleth.umich.edu/idp/Authn/RemoteUser?conversation=e1s1",
			"service":  "cosign-shibboleth.umich.edu",
			"required": "",
			"login":    username,
			"loginX":   username,
			"password": password,
		}).
		Post("https://weblogin.umich.edu/cosign-bin/cosign.cgi")
	if err != nil {
		panic(err)
	}
	if resp.StatusCode() != 200 {
		fmt.Println("cosgin post response not 200")
		return
	}
	str := NewSkipTillReader(strings.NewReader(resp.String()), []byte("TX|"))
	rtr := NewReadTillReader(str, []byte(":"))
	bs, err := ioutil.ReadAll(rtr)
	if err != nil {
		panic(err)
	}
	duo_config := string(bs)
	tx := duo_config[:len(duo_config)-1]
	resp, err = client.R().
		SetHeaders(map[string]string{
			"Host":            "api-d9c5afcf.duosecurity.com",
			"Referer":         "https://weblogin.umich.edu/",
			"Accept":          "*/*",
			"Accept-Encoding": "gzip, deflate, br",
			"Connection":      "keep-alive",
			"User-Agent":      "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/98.0.4758.102 Safari/537.36",
		}).
		Get(fmt.Sprintf("https://api-d9c5afcf.duosecurity.com/frame/web/v1/auth?tx=%s&parent=https%%3A%%2F%%2Fweblogin.umich.edu%%2Fcosign-bin%%2Fcosign.cgi&v=2.6", tx))
	if err != nil {
		panic(err)
	}
	client.SetCookies(resp.Cookies())
	// fmt.Println(fmt.Sprintf("https://api-d9c5afcf.duosecurity.com/frame/web/v1/auth?tx=%s&parent=https%%3A%%2F%%2Fweblogin.umich.edu%%2Fcosign-bin%%2Fcosign.cgi&v=2.6", tx))
	// ch_ua_brands := []string{" Not A;Brand", "Chromium", "Google Chrome"}
	resp, err = client.R().
		SetHeaders(map[string]string{
			"Referer":        fmt.Sprintf("https://api-d9c5afcf.duosecurity.com/frame/web/v1/auth?tx=%s&parent=https%%3A%%2F%%2Fweblogin.umich.edu%%2Fcosign-bin%%2Fcosign.cgi&v=2.6", tx),
			"Host":           "api-d9c5afcf.duosecurity.com",
			"Origin":         "https://api-d9c5afcf.duosecurity.com",
			"Sec-Fetch-Dest": "iframe",
			"Content-Type":   "application/x-www-form-urlencoded",
		}).
		SetFormData(map[string]string{
			"tx":                       tx,
			"parent":                   "https://weblogin.umich.edu/cosign-bin/cosign.cgi",
			"java_version":             "",
			"flash_version":            "",
			"screen_resolution_width":  "1536",
			"screen_resolution_height": "864",
			"color_depth":              "24",
			"ch_ua_brands":             "[\" Not A;Brand\",\"Chromium\",\"Google Chrome\"]",
			"ch_ua_mobile":             "false",
			"ch_ua_platform":           "Windows",
			"ch_ua_platform_version":   "14.0.0",
			"ch_ua_full_version":       "98.0.4758.102",
			"ch_ua_error":              "",
			"is_cef_browser":           "false",
			"is_ipad_os":               "false",
			"is_ie_compatibility_mode": "",
			"is_user_verifying_platform_authenticator_available":    "false",
			"user_verifying_platform_authenticator_available_error": "",
			"acting_ie_version": "",
			"react_support":     "true",
		}).
		Post(fmt.Sprintf("https://api-d9c5afcf.duosecurity.com/frame/web/v1/auth?tx=%s&parent=https%%3A%%2F%%2Fweblogin.umich.edu%%2Fcosign-bin%%2Fcosign.cgi&v=2.6", tx))
	if err != nil {
		panic(err)
	}
	client.SetCookies(resp.Cookies())
	fmt.Println(resp.String())

	str = NewSkipTillReader(strings.NewReader(resp.String()), []byte("name=\"sid\" value="))
	rtr = NewReadTillReader(str, []byte("\">"))
	bs, err = ioutil.ReadAll(rtr)
	if err != nil {
		panic(err)
	}
	cut_sid := string(bs)
	sid := strings.Replace(cut_sid, "&#x3d;", "=", 1)
	sid = strings.Replace(sid, "&#x7c;", "|", 3)
	sid = sid[18:]
	sid = sid[:len(sid)-2]

	str = NewSkipTillReader(strings.NewReader(resp.String()), []byte("name=\"_xsrf\" value=\""))
	rtr = NewReadTillReader(str, []byte("\" />"))
	bs, err = ioutil.ReadAll(rtr)
	if err != nil {
		panic(err)
	}
	cut_xsrf := string(bs)
	_xsrf := cut_xsrf[20:]
	_xsrf = _xsrf[:len(_xsrf)-4]
	fmt.Println(cut_xsrf)
	fmt.Println(_xsrf)

	resp, err = client.R().
		SetHeaders(map[string]string{
			"Host":             "api-d9c5afcf.duosecurity.com",
			"X-Xsrftoken":      _xsrf,
			"X-Requested-With": "XMLHttpRequest",
			"User-Agent":       "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/99.0.4844.51 Safari/537.36",
		}).
		SetFormData(map[string]string{
			"sid":              sid,
			"device":           "phone1",
			"factor":           "Duo Push",
			"out_of_date":      "",
			"days_out_of_date": "",
			"days_to_block":    "None",
		}).
		Post("https://api-d9c5afcf.duosecurity.com/frame/prompt")

	fmt.Println(resp.String())

	// if resp.StatusCode() != 302 {
	// 	fmt.Println("request for sid response not 302")
	// 	fmt.Println(resp.String())
	// 	fmt.Println(resp.Header())
	// 	return
	// }

	// sid_chunk := resp.Header().Get("location")
	// sid := NewSkipTillReader(strings.NewReader(resp.String()), []byte(sid_chunk))

	// resp, err = client.R().
	// 	Post("https://api-d9c5afcf.duosecurity.com/frame/prompt")
	// if err != nil {
	// 	panic(err)
	// }

	fmt.Println(username, password, url, location, description)
}
