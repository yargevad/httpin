package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"mime"
	"net/http"

	"github.com/gorilla/schema"
	"github.com/pkg/errors"
	"github.com/pressly/chi"
)

const (
	FormMultipart  = "multipart/form-data"
	FormURLEncoded = "application/x-www-form-urlencoded"
	JSONEncoded    = "application/json"
)

// App may contain handles that are safe for use by multiple goroutines.
type App struct {
	Form *schema.Decoder
}

// RequestType allows specifying the target type that a handler should
// unmarshal to, for supported Content-Types.
type RequestType interface {
	Allocate(app *App) interface{}
}

// Foo defines the field names that are legal as input for handlers
// using BodyParse with this type.
type Foo struct {
	App
	Name string
	ID   int
}

// FooHandler responds to POST requests for the path /foo
func FooHandler(w http.ResponseWriter, r *http.Request) {
	var foo *Foo
	foo = r.Context().Value("parsed-body").(*Foo)
	w.Write([]byte(fmt.Sprintf("%v", foo)))
}

// Allocate returns an initialized Foo, including an App reference.
func (f *Foo) Allocate(app *App) interface{} { return &Foo{App: *app} }

func main() {
	app := &App{Form: schema.NewDecoder()}
	router := chi.NewRouter()
	router.Use(app.BodyParse(&Foo{}))
	router.Post("/foo", FooHandler)

	log.Fatal(http.ListenAndServe(":8080", router))
}

// WriteResponse is a convenience method for returning from http handlers.
func WriteResponse(w http.ResponseWriter, code int, msg string) {
	w.WriteHeader(code)
	w.Write([]byte(msg))
}

type WrappedHandler func(h http.Handler) http.Handler

// BodyParse parses the POST body (using the provided RequestType)
// and stores the resulting struct pointer in the request context.
func (app *App) BodyParse(reqType RequestType) WrappedHandler {
	return func(h http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var inType string
			var err error
			var out interface{}

			if r.Method == "POST" {
				if inType, err = ContentType(r); err != nil {
					WriteResponse(w, 400, "couldn't parse media type")
					return
				}
				out = reqType.Allocate(app)
				switch inType {
				case JSONEncoded:
					defer r.Body.Close()
					if err = json.NewDecoder(r.Body).Decode(out); err != nil {
						WriteResponse(w, 400, "couldn't decode json")
						return
					}
				case FormURLEncoded:
					if err = r.ParseForm(); err != nil {
						WriteResponse(w, 400, "couldn't parse form")
						return
					}
					if err = app.Form.Decode(out, r.PostForm); err != nil {
						WriteResponse(w, 400, "couldn't decode form")
						return
					}
				default:
					WriteResponse(w, 400, "unsupported content-type")
					return
				}
				r = r.WithContext(context.WithValue(r.Context(), "parsed-body", out))
			}

			h.ServeHTTP(w, r)
		})
	}
}

// ContentType returns the MIME type of the input data, or an error,
// if the Content-Type header fails to parse.
func ContentType(r *http.Request) (mediaType string, err error) {
	contentType := r.Header.Get("Content-Type")
	if contentType != "" {
		mediaType, _, err = mime.ParseMediaType(contentType)
		if err != nil {
			return mediaType, errors.Wrap(err, "ParseMediaType failed")
		}
	}
	return mediaType, nil
}
