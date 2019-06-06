package templates

import (
	"html/template"
	"io/ioutil"
	"log"
	"path/filepath"
	"strings"
	"sync"
)

var cache *template.Template
var mutex sync.RWMutex

func Load(dir string) (err error) {
	mutex.Lock()
	defer mutex.Unlock()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		return err
	}
	cache = template.New("")

	for _, file := range files {
		rawdata, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			return err
		}
		data := string(rawdata)
		data = strings.ReplaceAll(data, "\n", "")
		data = strings.ReplaceAll(data, "\\n", "\n")
		tmpl := cache.New(file.Name())
		_, err = tmpl.Parse(data)
		if err != nil {
			return err
		}
	}

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
