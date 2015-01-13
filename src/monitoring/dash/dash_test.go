// Tests for dash.
package dash

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func init() {
	*templatePath = "tmpl/" // relative to test location
}

func TestIndexWithAuth(t *testing.T) {
	addPages()
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	indexPage.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("want status %d, got %q, with body %v\n", http.StatusOK, response.Code, response.Body)
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v\n", err)
	}
	sb := string(b)
	want := "signinButton"
	if !strings.Contains(sb, want) {
		t.Fatalf("want %q in response, didn't get it: %v [...]\n", want, sb[:200])
	}
}

func TestIndexNoAuth(t *testing.T) {
	addPages()
	*authDisabled = true
	request, _ := http.NewRequest("GET", "/", nil)
	response := httptest.NewRecorder()

	indexPage.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("want status %d, got %q, with body %v\n", http.StatusOK, response.Code, response.Body)
	}

	b, err := ioutil.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("failed to read body: %v\n", err)
	}
	sb := string(b)
	want := "Dashboard"
	if !strings.Contains(sb, want) {
		t.Fatalf("want %q in response, didn't get it: %v [...]\n", want, sb[200])
	}
}
