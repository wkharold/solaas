package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"os/exec"
	"text/template"
)

const (
	PROBLEM = "problem.jl"
)

type Optimization struct {
	Solver      string              `json:"solver"`
	Optimizer   string              `json:"optimizer"`
	Sense       string              `json:"sense"`
	Function    string              `json:"func"`
	Variables   []map[string]string `json:"variables"`
	Constraints []map[string]string `json:"constraints"`
}

var problemTemplate = `using JuMP, {{.Solver}}

m = Model(with_optimizer({{.Optimizer}}))


{{ range $index, $variable := .Variables }}@variable(m, {{ $variable.bounds }})
{{ end }}
@objective(m, {{.Sense}}, {{.Function}})

{{ range $index, $constraint := .Constraints }}@constraint(m, {{ $constraint.name }}, {{ $constraint.constraint }})
{{ end }}
JuMP.optimize!(m)

{{ range $index, $variable := .Variables }}println("{{ $variable.name }} = ", JuMP.value({{ $variable.name }}))
{{ end }}`

func toJulia(body io.ReadCloser) {
	b, err := ioutil.ReadAll(body)
	if err != nil {
		log.Fatal(err)
	}

	var opt Optimization
	err = json.Unmarshal(b, &opt)
	if err != nil {
		log.Fatal(err)
	}

	t, err := template.New("problem.jl").Parse(problemTemplate)
	if err != nil {
		log.Fatal(err)
	}

	f, err := os.Create(PROBLEM)
	if err != nil {
		log.Fatal(err)
	}

	t.Execute(f, opt)
}

func handler(w http.ResponseWriter, r *http.Request) {
	log.Print("Hello world received a request")

	toJulia(r.Body)

	cmd := exec.Command("julia", PROBLEM)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()
	if err != nil {
		log.Print("cmd.Run() failed")
	}

	fmt.Fprintf(w, "%s", string(stdout.Bytes()))
}

func main() {
	log.Print("Hello world sample started")

	http.HandleFunc("/", handler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Fatal(http.ListenAndServe(fmt.Sprintf(":%s", port), nil))
}
