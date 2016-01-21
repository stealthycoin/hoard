package horde

import (
	"html/template"
)


var (
	tMap = template.FuncMap{ "hoard": tFunction }
	nameToHash = map[string]string{}
	hashToName = map[string]string{}
	hoards = map[string]*HordeHandler{}
)


func addHoard(hh HordeHandler) {
	hoards[hh.Prefix] = &hh
}

func tFunction(in string) string {
	// Find matching hoard
	return preload(in)
}

func Funcs() template.FuncMap {
	return tMap
}
