package main

import (
	"context"
	"encoding/hex"
	"html"
	"html/template"
	"io/ioutil"
	"math/rand"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/HackDalton/coolcpu/cpu"
)

// careful! can only do concurrent reads, not writes
var templates map[string]*template.Template

type templateData struct {
	CPU    *cpu.CPU
	Err    error
	Output string
}

func writeTemplate(w http.ResponseWriter, template string, data templateData) {
	w.Header().Set("Content-Type", "text/html")
	err := templates[template].ExecuteTemplate(w, template+".tmpl", data)
	if err != nil {
		panic(err)
	}
}

func routeIndex(w http.ResponseWriter, r *http.Request) {
	writeTemplate(w, "index", templateData{})
}

func routeRun(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.PostFormValue("code") == "" {
		w.Write([]byte("You must enter in some code!"))
		return
	}

	codeStr := strings.Replace(r.PostFormValue("code"), " ", "", -1)

	code, err := hex.DecodeString(codeStr)
	if err != nil {
		w.Write([]byte("Code was not valid hex!<br><br>" + html.EscapeString(err.Error())))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	output := ""
	c := cpu.NewCPU()
	c.WriteCallback = func(c uint8) {
		output += string(c)
	}

	if len(code) > cpu.DefaultBankSize {
		w.Write([]byte("You've written too much code! The CoolCPU has a max of " + strconv.Itoa(cpu.DefaultBankSize) + "bytes of ROM."))
		return
	}

	for i := 0; i < len(code); i++ {
		c.ROM[i] = code[i]
	}

	copy(c.RAM[0x93-cpu.DefaultBankSize:], []byte("super secret flag!\x00"))

	err = c.Run(ctx)
	if err != nil {
		writeTemplate(w, "error", templateData{
			CPU: c,
			Err: err,
		})
		return
	}

	writeTemplate(w, "result", templateData{
		CPU:    c,
		Output: output,
	})
}

func main() {
	rand.Seed(time.Now().UnixNano())

	templates = map[string]*template.Template{}

	templateNames, err := ioutil.ReadDir("templates")
	if err != nil {
		panic(err)
	}
	for _, templateName := range templateNames {
		if templateName.Name() == "base.tmpl" {
			continue
		}

		cleanName := strings.Replace(templateName.Name(), ".tmpl", "", -1)
		templates[cleanName] = template.Must(template.ParseFiles(
			filepath.Join("templates", "base.tmpl"),
			filepath.Join("templates", templateName.Name()),
		))
	}

	http.Handle("/assets/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/run", routeRun)
	http.HandleFunc("/", routeIndex)

	err = http.ListenAndServe(":9485", nil)
	if err != nil {
		panic(err)
	}
}
