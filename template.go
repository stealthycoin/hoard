package hoard

import (
	"html/template"
)


var (
	tMap = template.FuncMap{
		"hoard": singleResource,
		"hoard_bundle": blockResources,
	}
	nameToHash = map[string]string{}
	hoards = map[string]*HoardHandler{}
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
