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