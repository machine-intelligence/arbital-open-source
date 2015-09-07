// Package pages provides some helpers for serving web pages.
//
// Example usage:
//   var myPage = pages.Add("/uri", myHandler, "tmpl/base.tmpl", "tmpl/page.tmpl")
//
//   func myHandler(w http.ResponseWriter, r *http.Request) pages.Result {
//     return pages.OK("some data to page.tmpl")
//   }
//
//   http.Handle(myPage.URI, myPage)
package pages

import (
	"fmt"
	"html/template"
	"net/http"
	"net/url"

	"zanaduu3/src/logger"
	"zanaduu3/src/user"
)

var (
	BaseTemplate        = "base"                                     // Name of top-level template to invoke for each page.
	BadRequestMsg       = "Invalid request. Please try again later." // Message to display if ShowError is called.
	StatusBadRequest    = Result{responseCode: http.StatusBadRequest}
	StatusUnauthorized  = Result{responseCode: http.StatusUnauthorized}
	StatusNotFound      = Result{responseCode: http.StatusNotFound}
	StatusInternalError = Result{responseCode: http.StatusInternalServerError}
)

// Renderer is a function to render a page result.
type Renderer func(w http.ResponseWriter, r *http.Request, u *user.User) *Result

// A Page to be rendered.
type Page struct {
	URI       string   // URI path
	Render    Renderer // func to render the page
	Templates []string // backing templates
}

// Add creates a new page.
//
// Add panics if the page templates cannot be parsed.
func Add(uri string, render Renderer, tmpls ...string) Page {
	return Page{
		URI:       uri,
		Render:    render,
		Templates: tmpls,
	}
}

// Result is the result of rendering a page.
type Result struct {
	data                interface{}      // Data to render the page.
	responseCode        int              // HTTP response code.
	err                 error            // Error, or nil.
	next                string           // Next uri, if applicable.
	funcMap             template.FuncMap // Functions map
	additionalTemplates []string         // Optional additional templates used to compile the page
}

// AddFuncMap adds the functions from the given funcMap to the functions
// already in the Result object.
func (r *Result) AddFuncMap(funcMap template.FuncMap) *Result {
	if r.funcMap == nil {
		r.funcMap = make(template.FuncMap)
	}
	for k, v := range funcMap {
		r.funcMap[k] = v
	}
	return r
}

// AddAdditionalTemplates adds more .tmpl files to be used in compiling the page.
func (r *Result) AddAdditionalTemplates(tmpls []string) *Result {
	r.additionalTemplates = append(r.additionalTemplates, tmpls...)
	return r
}

// StatusOK returns http.StatusOK with given data passed to the template.
func StatusOK(data interface{}) *Result {
	return &Result{
		responseCode: http.StatusOK,
		data:         data,
	}
}

// BadRequestWith returns a Result indicating a bad request.
func BadRequestWith(err error) *Result {
	return &Result{
		responseCode: http.StatusBadRequest,
		err:          err,
	}
}

// UnauthorizedWith returns a Result indicating an authorized request.
func UnauthorizedWith(err error) *Result {
	return &Result{
		responseCode: http.StatusUnauthorized,
		err:          err,
	}
}

// InternalErrorWith returns a Result indicating an internal error.
func InternalErrorWith(err error) *Result {
	return &Result{
		responseCode: http.StatusInternalServerError,
		err:          err,
	}
}

// RedirectWith returns a Result indicating to redirect to another URI.
func RedirectWith(uri string) *Result {
	return &Result{
		responseCode: http.StatusSeeOther,
		next:         uri,
	}
}

// CustomCodeWith returns a Result indicating the given responseCode.
func CustomCodeWith(data interface{}, responseCode int) *Result {
	return &Result{
		responseCode: responseCode,
		data:         data,
	}
}

// ShowError redirects to the index page with the "error" param set to
// a static error message.
//
// Provided error is logged, but not displayed to the user.
func ShowError(w http.ResponseWriter, r *http.Request, err error) {
	l := logger.GetLogger(r)
	q := url.Values{
		"error_msg": []string{BadRequestMsg},
	}
	nextUrl := fmt.Sprintf("/?%s", q.Encode())
	l.Errorf("returning StatusBadRequest and redirecting to %q: %v\n", nextUrl, err)
	http.Redirect(w, r, nextUrl, http.StatusSeeOther)
}

// Values are simple URL params.
type Values map[string]string

// UrlValues returns the simplifies values as url.Values.
func (vs Values) UrlValues() url.Values {
	q := url.Values{}
	for k, v := range vs {
		q[k] = []string{v}
	}
	return q
}

// AddTo adds the Values to specified URI.
func (v Values) AddTo(uri string) string {
	return fmt.Sprintf("%s?%s", uri, v.UrlValues().Encode())
}

// ServeHTTP serves HTTP for the page.
//
// ServeHTTP panics if no logger has been registered with SetLogger.
func (p Page) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	l := logger.GetLogger(r)
	l.Infof("Page %+v will ServeHTTP for URL: %v", p, r.URL)

	// Render the page, retrieving any data for the template.
	pr := p.Render(w, r, nil)
	if pr.err != nil || pr.responseCode != http.StatusOK {
		if pr.err != nil {
			l.Errorf("Error while rendering %v: %v\n", r.URL, pr.err)
		}
		if pr.responseCode == http.StatusNotFound {
			http.NotFound(w, r)
		} else if pr.responseCode == http.StatusBadRequest {
			http.Error(w, "Bad request", http.StatusBadRequest)
		} else if pr.responseCode == http.StatusSeeOther {
			http.Redirect(w, r, pr.next, http.StatusSeeOther)
		} else {
			http.Error(w, "Internal server error.", pr.responseCode)
		}
		return
	}

	allTemplates := append(p.Templates, pr.additionalTemplates...)
	template := template.Must(template.New(p.URI).Funcs(pr.funcMap).ParseFiles(allTemplates...))
	err := template.ExecuteTemplate(w, BaseTemplate, pr.data)
	if err != nil {
		// TODO: If this happens, partial template data is still written
		// to w by ExecuteTemplate, which isn't ideal; we'd like the 500
		// to be the only thing returned to viewing user.

		// Error rendering the template is a programming bug.
		l.Errorf("Failed to render template: %v", err)
		http.Error(w, "Internal server error.", http.StatusInternalServerError)
	}
}
