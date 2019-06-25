package templates

import (
	"html/template"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"

	"github.com/getsentry/sentry-go"
	"github.com/Masterminds/sprig"
)

func reportPanic(name string, data interface{}) {
	if perr := recover(); perr != nil {
		hub := sentry.CurrentHub().Clone()

		hub.Scope().SetExtra("template", name)
		hub.Scope().SetExtra("data", data)

		hub.Recover(perr)
	}
}

var cache *template.Template
var mutex sync.RWMutex

func Load(dir string) (err error) {
	mutex.Lock()
	defer mutex.Unlock()

	files, err := ioutil.ReadDir(dir)
	if err != nil {
		panic(err)
	}
	cache = template.New("").Funcs(sprig.FuncMap())

	for _, file := range files {
		rawdata, err := ioutil.ReadFile(filepath.Join(dir, file.Name()))
		if err != nil {
			panic(err)
		}
		data := string(rawdata)
		data = strings.ReplaceAll(data, "\n", "")
		data = strings.ReplaceAll(data, "\\n", "\n")
		tmpl := cache.New(file.Name())
		_, err = tmpl.Parse(data)
		if err != nil {
			panic(err)
		}
	}

	return err
}

func Execute(name string, data interface{}) (string, error) {
	defer reportPanic(name, data)

	mutex.RLock()
	defer mutex.RUnlock()

	var b strings.Builder

	err := cache.ExecuteTemplate(&b, name, data)
	if err != nil {
		panic(err)
	}

	return b.String(), nil
}

func Button(name string, data interface{}) string {
	defer reportPanic(name, data)

	mutex.RLock()
	defer mutex.RUnlock()

	var b strings.Builder

	err := cache.ExecuteTemplate(&b, name, data)
	if err != nil {
		panic(err)
	}

	return b.String()
}
