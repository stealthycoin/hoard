package hoard

import (
	"html/template"
)


var (
	tMap = template.FuncMap{ "hoard": tFunction }
	nameToHash = map[string]string{}
	hoards = map[string]*HoardHandler{}
)


func addHoard(hh *HoardHandler) {
	hoards[hh.Prefix] = hh
}

func tFunction(in string) string {
	// Find matching hoard
	return preload(in)
}

func Funcs() template.FuncMap {
	return tMap
}
