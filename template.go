package hoard

import (
	"fmt"
	"html/template"
)


var (
	tMap = template.FuncMap{
		"hoard": singleResource,
		"hoard_bundle": blockResources,
	}
	nameToHash = map[string]string{}
	hoards = map[string]*HoardHandler{}

	cssFmt = `<link rel="stylesheet" type="text/css" media="screen" href="%s" />`
	jsFmt = `<script type="text/javascript" src="%s"></script>`
)


func addHoard(hh *HoardHandler) {
	hoards[hh.Prefix] = hh
}

func singleResource(in string) string {
	// Find matching hoard
	return preload(in)
}

func blockResources(in ...string) (template.HTML, error) {
	// Load a block of resources into a single file
	return multiLoad(in)
}

func Funcs() template.FuncMap {
	return tMap
}


func wrapCSS(filename string) template.HTML {
	return template.HTML(fmt.Sprintf(cssFmt, filename))
}

func wrapJS(filename string) template.HTML {
	return template.HTML(fmt.Sprintf(jsFmt, filename))
}
