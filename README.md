# hoard
Golang static file server and cache library.


## Usage
#### Import
```
import "github.com/stealthycoin/hoard"
```

#### Linking a directory to a url prefix
```
hoard.Create("/static/", http.Dir("static"), []string{"text/css", "application/javascript"})
```
First argument is the url prefix for static files

Second argument is a directory to load files from

Third argument is a list of media types to compress