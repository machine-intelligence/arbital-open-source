// This package was taken from garybird/twitterstream package on github.
// It was modifid to support appengine.
package sessions

import (
	"bufio"
	"bytes"
	"compress/gzip"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
	"time"

	"golang.org/x/net/context"
	"google.golang.org/appengine/socket"

	"github.com/garyburd/go-oauth/oauth"
)

// HttpStream manages the connection to a Twitter streaming endpoint.
type HTTPStream struct {
	conn *tls.Conn
	r    *bufio.Scanner
	err  error
}

// HTTPStatusError represents an HTTP error return from the Twitter streaming
// API endpoint.
type HTTPStatusError struct {
	// HTTP status code.
	StatusCode int

	// Response body.
	Message string
}

func (err HTTPStatusError) Error() string {
	return "xelaie.HttpStream: status=" + strconv.Itoa(err.StatusCode) + " " + err.Message
}

var (
	responseLineRegexp = regexp.MustCompile("^HTTP/[0-9.]+ ([0-9]+) ")
	crlf               = []byte("\r\n")
)

// Open opens a new stream.
func OpenHTTPStream(c context.Context, oauthClient *oauth.Client, accessToken *oauth.Credentials, urlStr string, params url.Values) (*HTTPStream, error) {
	return openInternal(c, oauthClient, accessToken, urlStr, params)
}

func openInternal(c context.Context,
	oauthClient *oauth.Client,
	accessToken *oauth.Credentials,
	urlStr string,
	params url.Values) (*HTTPStream, error) {

	paramsStr := params.Encode()
	req, err := http.NewRequest("POST", urlStr, strings.NewReader(paramsStr))
	if err != nil {
		return nil, fmt.Errorf("1: %v", err)
	}

	req.Header.Set("Authorization", oauthClient.AuthorizationHeader(accessToken, "POST", req.URL, params))
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Content-Length", strconv.Itoa(len(paramsStr)))

	host := req.URL.Host
	port := "80"
	if h, p, err := net.SplitHostPort(req.URL.Host); err == nil {
		host = h
		port = p
	} else {
		if req.URL.Scheme == "https" {
			port = "443"
		}
	}

	ts := &HTTPStream{}
	socketConn, err := socket.DialTimeout(c, "tcp", host+":"+port, time.Minute)
	if err != nil {
		return nil, fmt.Errorf("2: %v", err)
	}

	if req.URL.Scheme == "https" {
		ts.conn = tls.Client(socketConn, &tls.Config{ServerName: host})
		if err := ts.conn.Handshake(); err != nil {
			return nil, ts.fatal(fmt.Errorf("3: %v", err))
		}
		if err := ts.conn.VerifyHostname(host); err != nil {
			return nil, ts.fatal(fmt.Errorf("4: %v", err))
		}
	}

	err = ts.conn.SetDeadline(time.Now().Add(60 * time.Second))
	if err != nil {
		return nil, ts.fatal(fmt.Errorf("5: %v", err))
	}

	if err := req.Write(ts.conn); err != nil {
		return nil, ts.fatal(fmt.Errorf("6: %v", err))
	}

	br := bufio.NewReader(ts.conn)
	resp, err := http.ReadResponse(br, req)
	if err != nil {
		return nil, ts.fatal(fmt.Errorf("7: %v", err))
	}

	if resp.Header.Get("Content-Encoding") == "gzip" {
		rc, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, ts.fatal(fmt.Errorf("8: %v", err))
		}
		resp.Body = rc
	}

	if resp.StatusCode != 200 {
		p, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, ts.fatal(fmt.Errorf("9: %v", err))
		}
		return nil, ts.fatal(HTTPStatusError{resp.StatusCode, string(p)})
	}

	ts.conn.SetWriteDeadline(time.Time{})

	ts.r = bufio.NewScanner(resp.Body)
	ts.r.Split(splitLines)
	return ts, nil
}

func (ts *HTTPStream) fatal(err error) error {
	if ts.conn != nil {
		ts.conn.Close()
	}
	if ts.err == nil {
		ts.err = err
	}
	return err
}

// Close releases the resources used by the stream. It can be called
// concurrently with Next.
func (ts *HTTPStream) Close() error {
	return ts.conn.Close()
}

// Err returns a non-nil value if the stream has a permanent error.
func (ts *HTTPStream) Err() error {
	return ts.err
}

// Next returns the next line from the stream. The returned slice is
// overwritten by the next call to Next.
func (ts *HTTPStream) Next() ([]byte, error) {
	if ts.err != nil {
		return nil, ts.err
	}
	for {
		// Twitter recommends reading with a timeout of 90 seconds.
		err := ts.conn.SetReadDeadline(time.Now().Add(90 * time.Second))
		if err != nil {
			return nil, ts.fatal(err)
		}
		if !ts.r.Scan() {
			err := ts.r.Err()
			if err == nil {
				err = io.EOF
			}
			return nil, ts.fatal(err)
		}
		p := ts.r.Bytes()
		if len(p) > 0 {
			return p, nil
		}
	}
}

// UnmarshalNext reads the next line of from the stream and decodes the line as
// JSON to data. This is a convenience function for streams with homogeneous
// entity types.
func (ts *HTTPStream) UnmarshalNext(data interface{}) error {
	p, err := ts.Next()
	if err != nil {
		return err
	}
	return json.Unmarshal(p, data)
}

func splitLines(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}
	if i := bytes.Index(data, crlf); i >= 0 {
		// We have a full CRLF terminated line.
		return i + 2, data[:i], nil
	}
	if atEOF {
		return 0, nil, io.ErrUnexpectedEOF
	}
	// Request more data.
	return 0, nil, nil
}
