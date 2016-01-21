package hoard

import (
	"bytes"
	"io"
	"io/ioutil"
	"log"
	"fmt"
	"path"
	"time"
	"mime"
	"net/http"
	"strings"
	"crypto/md5"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)


type FileBuffer struct {
	parent *HoardHandler
	buf []byte
}

func (fb *FileBuffer) Set(r io.Reader, ctype string) {
	// Check if it is a type that needs to be minified
	for _, t := range fb.parent.Types {
		if strings.HasPrefix(ctype, t) {
			// Found a matching mime type replace the reader
			if t == "application/javascript" {
				r = fb.parent.M.Reader("text/javascript", r)
			} else {
				r = fb.parent.M.Reader(t, r)
			}
			break
		}
	}

	// Read from the reader to our buffer
	buf, err := ioutil.ReadAll(r)
	if err != nil {
		log.Println(err)
	} else {
		fb.buf = make([]byte, len(buf))
		copy(fb.buf, buf)
	}
}


// Return a new reader to the file
func (fb *FileBuffer) Get() (*bytes.Reader, int) {
	reader := bytes.NewReader(fb.buf)
	return reader, len(fb.buf)
}


//
// A place for your dragon to store his treasures
//
type HoardHandler struct {
	Prefix  string
	Dir     http.Dir
	M       *minify.M
	Types   []string
	Stashed map[string]*FileBuffer
}


//
// Serve HTTP function to make it a handler interface
//
func (hh HoardHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// Remove the leading prefix
	urlPath := r.URL.Path[len(hh.Prefix):]


	var modTime time.Time
	// Open the path from the base diretory
	file, err := hh.Dir.Open(urlPath)
	if err != nil {
		// Could still be a hashed file
		modTime = time.Now()
	} else {
		// Get modification time
		fi, _ := file.Stat()
		modTime = fi.ModTime()
	}

	//Serve content from the file or from the cache
	if fb, ok := hh.Stashed[urlPath]; ok {
		// Serve from cached map
		content, _ := fb.Get()
		hh.ServeContent(w, r, urlPath, modTime, content)
	} else {
		// Serve the content from the file
		hh.ServeContent(w, r, urlPath, modTime, file)
	}
}


//
// Set the location of a hoard and register a handler with http module
//
func Create(prefix string, dir http.Dir, compress []string) {
	// Configure Minifier
	m := minify.New()
	for _, v := range compress {
		if v == "text/css" {
			m.AddFunc(v, css.Minify)
		} else if v == "application/javascript" {
			m.AddFunc("text/javascript", js.Minify)
		}
	}

	hh := HoardHandler{
		Prefix:  prefix,
		Dir:     dir,
		M:       m,
		Types:   compress,
		Stashed: make(map[string]*FileBuffer),
	}

	addHoard(&hh)
	http.Handle(prefix, hh)
}


//
// Preload a resource by adding it to the hoard
//
func preload(filePath string) string {
	// Find hoard it should belong to
	for key, h := range hoards {
		if strings.HasPrefix(filePath, key) {
			return addResource(filePath[len(key):], h)
		}
	}

	return filePath
}


//
// Add a resource to a hoard, or get its name if it already exists, returns the name
//
func addResource(name string, hh *HoardHandler) string {
	// Try to get the resource from the stash
	if _, ok := hh.Stashed[name]; ok {
		// It is already in the stash, return the hash for accessing it
		return nameToHash[name]
	} else {
		// Not in the stash, add it now since it will be requested once this page loads
		file, err := hh.Dir.Open(name)
		if err != nil {
			log.Println("Cannot find resource: ", name)
		}

		// Read file and get hash of contents
		fb := &FileBuffer{
			parent: hh,
			buf: make([]byte, 0),
		}
		fb.Set(file, mime.TypeByExtension(path.Ext(name)))
		hash := fmt.Sprintf("%x%s", md5.Sum(fb.buf), path.Ext(name))

		// Finally save to stash with both the real name and hashed name
		hh.Stashed[name] = fb
		hh.Stashed[hash] = fb

		nameToHash[name] = hh.Prefix + hash
		return hh.Prefix + hash
	}
}
