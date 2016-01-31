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


Hoard will now serve files such as ```website.com/static/js/main.js``` from the ```static``` directory.

## Template Usage

#### Load a single file

```
<link rel="stylesheet" type="text/css" media="screen" href="{{ hoard "/static/css/main.css" }}" />
```