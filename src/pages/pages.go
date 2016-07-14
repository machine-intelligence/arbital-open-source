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

	"zanaduu3/src/core"
	"zanaduu3/src/database"
	"zanaduu3/src/logger"
	"zanaduu3/src/sessions"
)

var (
	BaseTemplate        = "base"                                     // Name of top-level template to invoke for each page.
	BadRequestMsg       = "Invalid request. Please try again later." // Message to display if ShowError is called.
	StatusBadRequest    = Result{ResponseCode: http.StatusBadRequest}
	StatusUnauthorized  = Result{ResponseCode: http.StatusUnauthorized}
	StatusNotFound      = Result{ResponseCode: http.StatusNotFound}
	StatusInternalError = Result{ResponseCode: http.StatusInternalServerError}
)

// HandlerParams are passed to all handlers.
type HandlerParams struct {
	core.GlobalHandlerData

	W  http.ResponseWriter
	R  *http.Request
	C  sessions.Context
	DB *database.DB
	U  *core.CurrentUser
}

// Renderer is a function to render a page result. Returns:
// data - result data object. If nil, we have failed.
// string - error message we can display to the user
// error - error to display to the developers
type Renderer func(*HandlerParams) *Result

// A Page to be rendered.
type Page struct {
	URI       string   // URI path
	Render    Renderer // func to render the page
	Templates []string // backing templates
	Options   PageOptions
}

// PageOptions specify various requirements we need to check for the page.
// NOTE: make sure that default values are okay for all pages.
type PageOptions struct {
	AdminOnly    bool
	RequireLogin bool
	// If true, we don't care if the user is signed in or not.
	// This is used specifically for log-in and sign-up pages in private domains
	AllowAnyone bool
}

// Add creates a new page.
//
// Add panics if the page templates cannot be parsed.
func Add(uri string, render Renderer, options PageOptions, tmpls ...string) Page {
	return Page{
		URI:       uri,
		Render:    render,
		Templates: tmpls,
		Options:   options,
	}
}

// Result is the result of rendering a page.
type Result struct {
	Data                interface{}      // Data to render the page.
	ResponseCode        int              // HTTP response code.
	Err                 sessions.Error   // Error, or nil.
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
func Success(data interface{}) *Result {
	return &Result{
		ResponseCode: http.StatusOK,
		Data:         data,
	}
}

// StatusFailErr when the handler failed
func FailWith(err sessions.Error) *Result {
	return &Result{
		ResponseCode: http.StatusInternalServerError,
		Err:          err,
	}
}
func Fail(message string, err error) *Result {
	return FailWith(sessions.NewError(message, err))
}

// RedirectWith returns a Result indicating to redirect to another URI.
func RedirectWith(uri string) *Result {
	return &Result{
		ResponseCode: http.StatusSeeOther,
		next:         uri,
		Data:         uri,
	}
}

func (r *Result) Status(status int) *Result {
	r.ResponseCode = status
	return r
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
func (p Page) ServeHTTP(w http.ResponseWriter, r *http.Request, result *Result) {
	l := logger.GetLogger(r)
	l.Infof("Page %+v will ServeHTTP for URL: %v", p, r.URL)

	// Render the page, retrieving any data for the template.
	if result.Err != nil || result.ResponseCode != http.StatusOK {
		if result.Err != nil {
			l.Errorf("Error while rendering %v: %v\n", r.URL, result.Err)
		}
		if result.ResponseCode == http.StatusNotFound {
			http.NotFound(w, r)
		} else if result.ResponseCode == http.StatusBadRequest {
			http.Error(w, "Bad request", http.StatusBadRequest)
		} else if result.ResponseCode == http.StatusSeeOther {
			http.Redirect(w, r, result.next, http.StatusSeeOther)
		} else {
			http.Error(w, "Internal server error.", result.ResponseCode)
		}
		return
	}

	allTemplates := append(p.Templates, result.additionalTemplates...)
	template := template.Must(template.New(p.URI).Funcs(result.funcMap).ParseFiles(allTemplates...))
	err := template.ExecuteTemplate(w, BaseTemplate, result.Data)
	if err != nil {
		// TODO: If this happens, partial template data is still written
		// to w by ExecuteTemplate, which isn't ideal; we'd like the 500
		// to be the only thing returned to viewing user.

		// Error rendering the template is a programming bug.
		l.Errorf("Failed to render template: %v", err)
		http.Error(w, "Internal server error.", http.StatusInternalServerError)
	}
}
