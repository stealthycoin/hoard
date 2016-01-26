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
	"errors"
	"crypto/md5"
	"html/template"

	"github.com/tdewolff/minify"
	"github.com/tdewolff/minify/css"
	"github.com/tdewolff/minify/js"
)


type FileBuffer struct {
	parent *HoardHandler // Handler responsiblef for this buffer

	// Only one of the following should be set
	buf    []byte        // This filebuffer has its own content
	deps   []*FileBuffer // This filebuffer is a collection of other filebuffers
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


// Return a new reader to the buffer content
func (fb *FileBuffer) Get() (io.ReadSeeker, int) {
	if fb.deps != nil {
		// Dependant file buffer with no real content of its own
		// Collect readers from all dependencides in order
		readers := make([]io.ReadSeeker, 0)
		tot_len := 0
		for _, dep := range fb.deps {
			reader, length := dep.Get()
			readers = append(readers, reader)
			tot_len += length
		}

		// Build a multireader from the readers
		return MultiReadSeeker(readers...), tot_len
	}

	// Otherwise read out the data from the buffer
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
// Remove the shared prefix of a particular hoard
//
func (hh *HoardHandler) RemovePrefix(in string) string {
	return in[len(hh.Prefix):]
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
	if _, ok := hh.Stashed[urlPath]; !ok {
		// Add the file if its not found
		addResource(urlPath, &hh)
	}

	// Serve from cached map
	fb, _ := hh.Stashed[urlPath]
	content, _ := fb.Get()
	hh.ServeContent(w, r, urlPath, modTime, content)
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
			buf:    make([]byte, 0),
			deps:   nil,
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


//
// Add a block of resources
//
func multiLoad(names []string) (template.HTML, error) {
	// Verify we have multiple files
	if len(names) == 0 {
		return "", errors.New("hoard_bundle tag with no filenames.")
	} else if len(names) == 1 {
		return "", errors.New("hoard_bundle tag with 1 file. Use the haord tag instead.")
	}

	// Housekeeping
	buffers := make([]*FileBuffer, 0)
	ctype := ""
	ext := ""
	var hh *HoardHandler

	// Find hoard this block belongs to using the first filename
	for key, h := range hoards {
		if strings.HasPrefix(names[0], key) {
			hh = h
		}
	}


	// Concat names and get MD5
	longName := strings.Join(names, "")
	hash := fmt.Sprintf("%x%s", md5.Sum([]byte(longName)), path.Ext(names[0]))

	// Check for existing hash, if it exists return early
	if _, ok := hh.Stashed[hash]; ok {
		if ext == ".css" {
			return wrapCSS(hash), nil
		} else if ext == ".js" {
			return wrapJS(hash), nil
		}
	}

	// It doesn't exist, we need to collect all the files and mark them as dependencies
	// Go through each name
	for _, name := range names {
		// Find hoard it should belong to
		ext = path.Ext(name)
		temp := mime.TypeByExtension(ext)

		// Verify it matches previous filetypes
		if temp != ctype && ctype != "" {
			return "", errors.New("Mismatched mimetype in hoard_bundle")
		} else {
			ctype = temp
		}

		// Get the hash of this dependency (load it if unloaded)
		dep_hash := addResource(hh.RemovePrefix(name), hh)

		// Add it to the list of dependencies
		if new_dep, ok := hh.Stashed[hh.RemovePrefix(dep_hash)]; ok {
			buffers = append(buffers, new_dep)
		} else {
			log.Println("Failed to load dependency", dep_hash)
		}
	}

	// Read file(s) and get hash of contents
	fb := &FileBuffer{
		parent: hh,
		buf:    nil,
		deps:   buffers,
	}


	// Save under hash only since this isnt a single file
	hh.Stashed[hash] = fb

	// Need to surround it in its tag
	result := hh.Prefix + hash
	if ext == ".css" {
		return wrapCSS(result), nil
	} else if ext == ".js" {
		return wrapJS(result), nil
	}

	return template.HTML(result), nil
}
