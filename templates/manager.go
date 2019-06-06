package templates

import (
	"log"
	"path/filepath"
	"strings"
	"sync"
	"text/template"
)

var cache *template.Template
var mutex sync.RWMutex

func Load(dir string) (err error) {
	mutex.Lock()
	defer mutex.Unlock()

	pattern := filepath.Join(dir, "*.tmpl")

	cache, err = template.ParseGlob(pattern)

	return err
}

func Execute(name string, data interface{}) (string, error) {
	mutex.RLock()
	defer mutex.RUnlock()

	var b strings.Builder

	err := cache.ExecuteTemplate(&b, name, data)
	if err != nil {
		return "", err
	}

	return b.String(), nil
}

func Button(name string, data interface{}) string {
	mutex.RLock()
	defer mutex.RUnlock()

	var b strings.Builder

	err := cache.ExecuteTemplate(&b, name, data)
	if err != nil {
		log.Println(err)
		return ""
	}

	return b.String()
}
