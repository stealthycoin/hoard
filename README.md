# Hoard
Golang static file server and cache library.

Hoard can be used to serve/compress/cache static files as well as serve groups of files as a single file.


## Usage
#### Import
```
import "github.com/stealthycoin/hoard"
```

#### Linking a directory to a url prefix
```
hoard.Create("/static/", http.Dir("static"), []string{"text/css", "application/javascript"})
```
First argument is the url prefix for static files.
Second argument is a directory to load files from.
Third argument is a list of media types to compress.


Hoard will now serve files such as ```website.com/static/js/main.js``` from the ```static``` directory.

## Template Usage

#### Load a single file

```
<link rel="stylesheet" type="text/css" media="screen" href="{{ hoard "/static/css/main.css" }}" />
```

```
<script type="text/javascript" src="{{ hoard "/static/js/main.js" }}"></script>
```

#### Load multiple files at once

```
{{ hoard_bundle "/static/js/main.js" "/static/js/pageone.js" "/static/js/secondary.js" }}
```

Hoard will load these three files sequentially together into one and compress them (if it's set to) and return them as a single file. This will decrease the number of files the browser needs to request. Hoard also caches all files. It will also wrap it in either a js script tag or link tag for css. If the file extensions do not match it will throw an error. If the files are not css or js just the filename will be returned.