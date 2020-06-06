package main

import (
	"context"
	"encoding/hex"
	"html"
	"html/template"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/HackDalton/coolcpu/cpu"
)

// careful! can only do concurrent reads, not writes
var templates map[string]*template.Template

var version cpu.Version

type templateData struct {
	CPU     *cpu.CPU
	Version cpu.Version
	Err     error
	Output  string
}

func writeTemplate(w http.ResponseWriter, template string, data templateData) {
	w.Header().Set("Content-Type", "text/html")
	err := templates[template].ExecuteTemplate(w, template+".tmpl", data)
	if err != nil {
		panic(err)
	}
}

func routeIndex(w http.ResponseWriter, r *http.Request) {
	writeTemplate(w, "index", templateData{
		Version: version,
	})
}

func routeRun(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/html")

	if r.PostFormValue("code") == "" {
		w.Write([]byte("You must enter in some code!"))
		return
	}

	codeStr := strings.ReplaceAll(strings.ReplaceAll(strings.ReplaceAll(r.PostFormValue("code"), " ", ""), "\n", ""), "\r", "")

	code, err := hex.DecodeString(codeStr)
	if err != nil {
		w.Write([]byte("Code was not valid hex!<br><br>" + html.EscapeString(err.Error())))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	output := ""
	c := cpu.NewCPU(version)
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

	if c.Version == cpu.Version1 {
		copy(c.RAM[0x93-cpu.DefaultBankSize:], []byte("hackDalton{l00p_d3_l00p_syYBaqvCvi}\x00"))
	}

	err = c.Run(ctx)
	if err != nil {
		templateName := "error"
		if err == context.DeadlineExceeded {
			templateName = "timeout"
		}

		writeTemplate(w, templateName, templateData{
			CPU:     c,
			Version: version,
			Err:     err,
			Output:  output,
		})
		return
	}

	writeTemplate(w, "result", templateData{
		CPU:     c,
		Version: version,
		Output:  output,
	})
}

func main() {
	rand.Seed(time.Now().UnixNano())

	if len(os.Args) != 2 {
		log.Println("You must set the CPU version to use (either 1 or 2)!")
		log.Fatalln("Usage: coolcpu [version]")
	}

	versionString := os.Args[1]
	if versionString == "1" {
		version = cpu.Version1
	} else if versionString == "2" {
		version = cpu.Version2
	} else {
		log.Fatalf("Invalid version '%s'.", versionString)
	}

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
		templates[cleanName] = template.Must(
			template.New(templateName.Name()).Funcs(template.FuncMap{
				"hex": func(i uint8) string {
					return strconv.FormatInt(int64(i), 16)
				},
				"hexdump": func(data [cpu.DefaultBankSize]uint8) string {
					return hex.Dump(data[:])
				},
			}).ParseFiles(
				filepath.Join("templates", "base.tmpl"),
				filepath.Join("templates", templateName.Name()),
			),
		)
	}

	http.Handle("/assets/", http.FileServer(http.Dir("./")))
	http.HandleFunc("/run", routeRun)
	http.HandleFunc("/", routeIndex)

	err = http.ListenAndServe(":9485", nil)
	if err != nil {
		panic(err)
	}
}
