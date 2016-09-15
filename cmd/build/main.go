// Copyright 2015-2016 Sevki <s@sevki.org>. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main // import "sevki.org/build/cmd/build"

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"sevki.org/build"
	"sevki.org/build/builder"
	_ "sevki.org/build/targets/build"
	_ "sevki.org/build/targets/cc"
	_ "sevki.org/build/targets/harvey"
	_ "sevki.org/build/targets/yacc"
	"sevki.org/build/term"
	"sevki.org/build/util"
	"sevki.org/lib/prettyprint"
)

var (
	buildVer = "version"
)
var (
	verbose    = flag.Bool("v", false, "more verbose output")
	andInstall = flag.Bool("i", false, "install after a build")
)

func main() {
	log.SetFlags(log.LstdFlags | log.Lshortfile)
	flag.Usage = Usage
	flag.Parse()

	if util.GetProjectPath() == "" {
		Usage()
	}
	target := flag.Arg(0)
	switch target {
	case "version":
		version()
		return
	case "nuke":
		os.RemoveAll("/tmp/build")
		if len(flag.Args()) >= 2 {
			target = flag.Args()[1]
			execute(target)
		}
	case "query":
		target = flag.Args()[1]
		query(target)
	case "hash":
		target = flag.Args()[1]
		hash(target)
	case "install":
		if err := install(); err != nil {
			log.Fatal(err)
		}
	case "mk":
		fallthrough
	default:
		execute(target)
		if *andInstall {
			if err := install(); err != nil {
				log.Fatal(err)
			}
		}
	}
}

func Usage() {
	fmt.Fprintf(os.Stderr, "Usage of %s:\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "\t%s [-iv] [command] target\n\n", os.Args[0])
	flag.PrintDefaults()
	fmt.Fprintln(os.Stderr, "\nCommands:")
	fmt.Fprintln(os.Stderr, "version\n\tPrint version and exit.")
	fmt.Fprintln(os.Stderr, "nuke\n\tRemove the cache directory.")
	fmt.Fprintln(os.Stderr, "query\n\tPrint a json representation of the target.")
	fmt.Fprintln(os.Stderr, "hash\n\tPrint the hash of the target.")
	fmt.Fprintln(os.Stderr, "install\n\tCopy outputs into the project tree.")
	fmt.Fprintln(os.Stderr, "mk (default)\n\tBuild the target.")

	// general help
	fmt.Println(`
We require that you run this application inside a git repository.
All packages are "relative" to the top-level of the git repository,
which is spelled "//" in targets.

Note that an empty target will build the default target of the current
package.

Targets are of the form "//<package>:<rule>".

If the target doesn't start with "//", it is interpreted as relative to
the working directory.  If the rule is empty, the last element of the
package is used.  Relative targets may omit the leading colon.
`)

	os.Exit(1)
}

func progress() {
	fmt.Println(runtime.NumCPU())
}

func version() {
	fmt.Printf("Build %s\n", buildVer)
	os.Exit(0)
}

func doneMessage(s string) {
	fmt.Printf("[%s] %s\n", " OK ", s)
}

func failMessage(s string) {
	fmt.Printf("[ %s ] %s\n", "FAIL", s)
}

func hash(t string) {
	c := builder.New()
	fmt.Printf("%x\n", c.Add(t).HashNode())
}

func query(t string) {
	c := builder.New()
	fmt.Println(prettyprint.AsJSON(c.Add(t).Target))
}

func execute(t string) {
	c := builder.New()

	c.Root = c.Add(t)
	c.Root.IsRoot = true

	if c.Root == nil {
		log.Fatal("We couldn't find the root")
	}

	cpus := int(float32(runtime.NumCPU()) * 1.25)

	done := make(chan bool)

	// If the app hangs, there is a log.
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, os.Interrupt)
	go func() {
		<-sigs
		f, _ := os.Create("/tmp/build-crash-log.json")
		fmt.Fprintf(f, prettyprint.AsJSON(c.Root))
		os.Exit(1)
	}()

	go term.Listen(c.Updates, cpus, *verbose)
	go term.Run(done)

	go c.Execute(time.Second, cpus)
	for {
		select {
		case done := <-c.Done:
			if *verbose {
				doneMessage(done.Url.String())
			}
			if done.IsRoot {
				goto FIN
			}
		case err := <-c.Error:
			<-done

			log.Fatal(err)
			os.Exit(1)
		case <-c.Timeout:
			log.Println("your build has timed out")
		}

	}
FIN:
	term.Exit()
	<-done
}

func makeGraph(n *builder.Node, g *Graph) {
	if _, exists := g.Targets[n.Url.String()]; !exists {
		g.Targets[n.Url.String()] = n.Target
		g.Outputs[n.Url.String()] = n.Output
		for _, c := range n.Children {
			makeGraph(c, g)
		}
	}

}

type Graph struct {
	Root    *builder.Node
	Outputs map[string]string
	Targets map[string]build.Target
}

func compare(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i, c := range a {
		if c != b[i] {
			return false
		}
	}
	return true
}

// Copy everything from $BUILD_OUT into "//"
func install() error {
	p := util.GetProjectPath()
	return filepath.Walk(util.BuildOut(), mvTo(p))
}

func mvTo(root string) filepath.WalkFunc {
	out := util.BuildOut()

	return func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		rel := strings.TrimPrefix(path, out)
		if rel == "" {
			return nil
		}
		to := filepath.Join(root, rel)

		if info.IsDir() {
			if _, err := os.Stat(to); err != nil {
				return os.Mkdir(to, 0755)
			}
			return nil
		}

		new, err := os.OpenFile(to+"_", os.O_WRONLY|os.O_CREATE, info.Mode())
		if err != nil {
			failMessage(rel)
			return err
		}
		defer new.Close()
		built, err := os.Open(path)
		if err != nil {
			failMessage(rel)
			return err
		}
		defer built.Close()
		if _, err := io.Copy(new, built); err != nil {
			failMessage(rel)
			return err
		}

		if *verbose {
			doneMessage(rel)
		}
		return os.Rename(to+"_", to)
	}
}
