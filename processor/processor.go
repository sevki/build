// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package processor

import (
	"bytes"
	"encoding/json"
	"fmt"
	"path/filepath"
	"reflect"

	"log"

	"os"
	"os/exec"
	"regexp"

	"strings"

	"bldy.build/build"
	"bldy.build/build/ast"
	"bldy.build/build/internal"
	"bldy.build/build/parser"
	"bldy.build/build/preprocessor"
	"bldy.build/build/util"
)

type Processor struct {
	vars    map[string]interface{}
	wd      string
	seen    map[string]*ast.Func
	parser  *parser.Parser
	Targets chan build.Target
}

func NewProcessor(p *parser.Parser) *Processor {
	return &Processor{
		vars:    make(map[string]interface{}),
		parser:  p,
		Targets: make(chan build.Target),
		seen:    make(map[string]*ast.Func),
	}
}
func NewProcessorFromURL(url parser.TargetURL, wd string) (*Processor, error) {

	BUILDPATH := filepath.Join(url.BuildDir(wd, util.GetProjectPath()), "BUILD")
	BUCKPATH := filepath.Join(url.BuildDir(wd, util.GetProjectPath()), "BUCK")

	var fp string

	if _, err := os.Stat(BUCKPATH); err == nil {
		fp = BUCKPATH
	} else if _, err := os.Stat(BUILDPATH); err == nil {
		fp = BUILDPATH
	} else {
		return nil, err
	}
	return NewProcessorFromFile(fp)

}

func NewProcessorFromFile(n string) (*Processor, error) {

	ks, err := os.Open(n)
	if err != nil {
		return nil, fmt.Errorf("opening file: %s\n", err.Error())
	}
	ts, _ := filepath.Abs(ks.Name())
	dir := strings.Split(ts, "/")
	p := parser.New(n, "/"+filepath.Join(dir[:len(dir)-1]...), ks)

	return NewProcessor(p), nil
}

func (p *Processor) Run() {

	go p.parser.Run()
	var d ast.Decl

	// Define a set of preprocessors
	preprocessors := []preprocessor.PreProcessor{
		&preprocessor.DuplicateLoadChecker{
			Seen: make(map[string]*ast.Func),
		},
	}

	for d = <-p.parser.Decls; d != nil; d = <-p.parser.Decls {
		// Run preprocessors
		for _, pp := range preprocessors {
			var err error
			d, err = pp.Process(d)
			if err != nil {
				log.Fatal(err)
			}
		}

		switch d.(type) {
		case *ast.Error:
			log.Printf(d.(*ast.Error).Error.Error())
		case *ast.Func:
			p.runFunc(d.(*ast.Func))
		case *ast.Assignment:
			p.doAssignment(d.(*ast.Assignment))
		case *ast.Loop:
			p.doLoop(d.(*ast.Loop))
		default:
			//			log.Printf("%T", d)
		}
	}
	p.Targets <- nil
}
func (p *Processor) doLoop(l *ast.Loop) {
	_range := p.unwrapValue(l.Range)
	if _range == nil {
		return
	}
	var tmp interface{}
	exists := false

	tmp, exists = p.vars[l.Key]

	for _, v := range _range.([]interface{}) {
		p.vars[l.Key] = v
		p.runFunc(l.Func)

	}
	if exists {
		p.vars[l.Key] = tmp
	} else {
		delete(p.vars, l.Key)

	}
}
func (p *Processor) doAssignment(a *ast.Assignment) {
	p.vars[a.Key] = p.unwrapValue(a.Value)
}
func (p *Processor) unwrapFunc(f *ast.Func) *ast.Func {
	nf := *f
	nf.Params = p.unwrapMap(f.Params)
	nf.AnonParams = p.unwrapSlice(f.AnonParams)
	return &nf
}
func (p *Processor) unwrapSlice(slc []interface{}) (ns []interface{}) {
	for _, v := range slc {
		t := p.unwrapValue(v)
		if t != nil {
			ns = append(ns, t)
		} else {
			log.Fatalf("unwrapping of value %v failed.", v)
		}
	}
	return ns
}

func (p *Processor) unwrapMap(mp map[string]interface{}) (nm map[string]interface{}) {
	nm = make(map[string]interface{})
	for k, v := range mp {
		nm[k] = p.unwrapValue(v)
	}
	return nm
}
func (p *Processor) unwrapValue(i interface{}) interface{} {
	switch i.(type) {
	case *ast.BasicLit:
		return i.(*ast.BasicLit).Interface()
	case *ast.Variable:
		if v, ok := p.vars[i.(*ast.Variable).Key]; ok {
			return v
		} else {
			log.Fatalf("variable %s is not present in %s. make sure it's loaded properly or declared", i.(*ast.Variable).Key, p.parser.Path)
		}
		return nil
	case *ast.Slice:
		return p.unwrapSlice(i.(*ast.Slice).Slice)
	case *ast.Map:
		return p.unwrapMap(i.(*ast.Map).Map)
	case *ast.Func:
		return p.funcReturns(i.(*ast.Func))
	default:
		return nil
	}
}
func (p *Processor) runFunc(f *ast.Func) {
	f = p.unwrapFunc(f)
	switch f.Name {
	case "load":
		fail := func() {
			log.Fatal("should be used like so; load(file, var...)")
		}

		filePath := ""
		var varsToImport []string
		// Check paramter types
		for i, param := range f.AnonParams {
			switch param.(type) {
			case string:
				v := param.(string)
				if i == 0 {
					filePath = v
				} else {
					varsToImport = append(varsToImport, v)
				}
				break
			default:
				fail()
			}
		}
		loadingProcessor, err := NewProcessorFromFile(p.absPath(filePath))
		if err != nil {
			log.Fatal(err)
		}
		go loadingProcessor.Run()

		for d := <-loadingProcessor.Targets; d != nil; d = <-loadingProcessor.Targets {
		}
		if err != nil {
			log.Fatal(err)
		}
		if p.vars == nil {
			p.vars = make(map[string]interface{})
		}

		for _, v := range varsToImport {
			if val, ok := loadingProcessor.vars[v]; ok {
				p.vars[v] = val
			} else {
				log.Fatalf("%s is not present at %s. Please check the file and try again.", v, filePath)
			}
		}

	default:
		targ, err := p.makeTarget(f)
		if err != nil {
			log.Fatal(err)
			return
		} else {
			p.Targets <- targ
		}
	}
}

func (p *Processor) absPath(s string) string {
	var r string
	if strings.TrimLeft(s, "//") != s {
		r = filepath.Join(util.GetProjectPath(), strings.Trim(s, "//"))
	} else {
		r = filepath.Join(p.parser.Path, s)
	}
	r = os.Expand(r, util.Getenv)
	return r
}

func (p *Processor) makeTarget(f *ast.Func) (build.Target, error) {

	if v, ok := p.vars[f.Name]; ok {
		switch v.(type) {
		case *ast.Func:

			macro := v.(*ast.Func)
			f.Name = macro.Name
			for k, v := range macro.Params {
				if _, ok := f.Params[k]; !ok {
					f.Params[k] = v
				}
			}
		}
	}
	ttype := internal.Get(f.Name)

	payload := make(map[string]interface{})

	for key, fn := range f.Params {

		field, err := internal.GetFieldByTag(f.Name, key, ttype)
		if err != nil {
			return nil, err
		}

		var i interface{}
		switch fn.(type) {
		case *ast.Func:
			x := fn.(*ast.Func)
			i = p.funcReturns(x)
		case ast.Variable:
			i = p.vars[fn.(ast.Variable).Key]
		default:
			i = fn
		}

		if field.Type != reflect.TypeOf(i) {
			// return nil, fmt.Errorf("%s is of type %s not %s.", key, reflect.TypeOf(i).String(), field.Type.String())
		}

		payload[field.Name] = i
		if key == "name" {
			name := i.(string)
			if exst, ok := p.seen[name]; ok {
				dupeErr := `Target %s is declared more than once at these locations:
	 %s:%d: 
	 %s:%d: `

				return nil, fmt.Errorf(dupeErr, name, f.File, f.Start.Line, exst.File, exst.Start.Line)
			} else {
				p.seen[name] = f
			}
		}
	}

	//BUG(sevki): this is a very hacky way of doing this but it seems to be safer.
	var bytz []byte
	buf := bytes.NewBuffer(bytz)

	enc := json.NewEncoder(buf)
	enc.Encode(payload)

	t := reflect.New(ttype).Interface()
	dec := json.NewDecoder(buf)
	dec.Decode(t)
	switch t.(type) {
	case build.Target:
		break
	default:
		log.Fatalf("type %s doesn't implement the build.Target interface, check sevki.co/2LLRfc for more information", ttype.String())
	}
	return t.(build.Target), nil
}

func (p *Processor) funcReturns(f *ast.Func) interface{} {
	f = p.unwrapFunc(f)

	switch f.Name {
	case "glob":
		return p.glob(f)
	case "version":
		return p.version(f)
	case "addition":
		return p.combineArrays(f)
	case "slice":
		return p.sliceArray(f)
	case "index":
		return p.indexArray(f)
	case "env":
		return p.env(f)
	default:
		return f
	}
}

func (p *Processor) combineArrays(f *ast.Func) interface{} {
	var t []interface{}

	for _, v := range f.AnonParams {
		switch v.(type) {
		case []interface{}:
			t = append(t, v.([]interface{})...)
		}
	}

	return t
}

func (p *Processor) indexArray(f *ast.Func) interface{} {
	index, hasIndex := f.Params["index"].(int)

	if !hasIndex {
		return nil
	}

	switch f.Params["var"].(type) {
	case []interface{}:
		return f.Params["var"].([]interface{})[index]
	}
	return nil
}
func (p *Processor) sliceArray(f *ast.Func) interface{} {
	switch f.Params["var"].(type) {
	case []interface{}:
		return p.sliceInterfaceArray(f, f.Params["var"].([]interface{}))
	case string:
		return p.sliceString(f, f.Params["var"].(string))
	}

	return nil
}

func (p *Processor) sliceInterfaceArray(f *ast.Func, s []interface{}) interface{} {
	start, hasStart := f.Params["start"].(int)
	end, hasEnd := f.Params["end"].(int)
	switch {
	case hasStart && hasEnd:
		return s[start:end]
	case hasStart:
		return s[start:]
	case hasEnd:
		return s[:end]
	default:
		return nil
	}
}

func (p *Processor) sliceString(f *ast.Func, s string) interface{} {
	start, hasStart := f.Params["start"].(int)
	end, hasEnd := f.Params["end"].(int)
	if end < 0 {
		end = len(s) + end
	}

	switch {
	case hasStart && hasEnd:
		return s[start:end]
	case hasStart:
		return s[start:]
	case hasEnd:
		return s[:end]
	default:
		return nil
	}
}
func (p *Processor) glob(f *ast.Func) []string {
	wd := p.parser.Path
	if !filepath.IsAbs(wd) {
		log.Fatalf("Error parsing glob: %s is not an absolute path.", wd)
	}

	var files []string
	var excludes []*regexp.Regexp

	if len(f.AnonParams) != 1 {
		return []string{"Error parsing glob: proper usage is like so glob(include, exclude=[], exclude_directories=1)"}
	}

	if exs, ok := f.Params["exclude"]; ok {

		for _, ex := range exs.([]interface{}) {
			r, _ := regexp.Compile(ex.(string))
			excludes = append(excludes, r)
		}
	}

	//BUG(sevki): put some type checking here
	for _, s := range f.AnonParams[0].([]interface{}) {
		globPtrn := ""

		switch s.(type) {
		case string:
			globPtrn = filepath.Clean(filepath.Join(wd, s.(string)))
			log.Println(globPtrn)
		default:
			return nil
		}

		globFiles, err := filepath.Glob(globPtrn)

		if err != nil {
			return []string{"Error parsing glob: %s"}
		}

		for _, f := range globFiles {
			f, _ := filepath.Rel(util.GetProjectPath(), f)
			f = fmt.Sprintf("//%s", f)
		}
	RESIZED:
		for i, f := range globFiles {
			for _, x := range excludes {
				if x.Match([]byte(f)) {
					globFiles = append(globFiles[:i], globFiles[i+1:]...)
					goto RESIZED
				}
			}
			globFiles[i] = f
		}

		files = append(files, globFiles...)
	}
	return files
}

func (p *Processor) env(f *ast.Func) string {
	if len(f.AnonParams) != 1 {
		return ""
	}
	return util.Getenv(f.AnonParams[0].(string))
}

func (p *Processor) version(f *ast.Func) string {

	if out, err := exec.Command("git",
		"--git-dir="+util.GetGitDir(p.parser.Path)+".git",
		"describe",
		"--always").Output(); err != nil {
		return err.Error()
	} else {
		return strings.TrimSpace(string(out))
	}
}
