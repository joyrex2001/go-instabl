package main

import (
	"fmt"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// main will take you to the bat-mobile and yell "let's go!"
func main() {
	if len(os.Args) != 2 {
		fmt.Fprintf(os.Stderr, "usage: %s <repository>\n", os.Args[0])
		os.Exit(1)
	}

	anlzr, err := NewAnalyzer(os.Args[1])
	if err != nil {
		fmt.Fprintf(os.Stderr, "analyze failed: %s\n", err)
		os.Exit(2)
	}

	anlzr.Analyze()
	Report(anlzr.GetStats())
}

// Report will output the metrics to the console.
func Report(stats Stats) {
	lst := []string{}
	for nm, st := range stats {
		lst = append(lst, fmt.Sprintf("%.2f\t%s", st.Instability(), nm))
	}
	sort.Strings(lst)
	fmt.Println(strings.Join(lst, "\n"))
}

// Stats is a hashmap that contains the stability metrics for all collected
// packages.
type Stats map[string]Stability

// Stability is the structure that keeps track if incoming dependencies (FanIn),
// and outgoing dependencies (FanOut)
type Stability struct {
	FanIn  int
	FanOut int
}

// Instability will calculate the instability metric.
func (s Stability) Instability() float64 {
	if s.FanIn+s.FanOut == 0 {
		return .0
	}
	return float64(s.FanOut) / (float64(s.FanIn) + float64(s.FanOut))
}

// Analyzer is the object that will analyze a provided repo, collecting the
// dependency statistics.
type Analyzer struct {
	stats Stats
	repo  string
}

// NewAnalyzer will instantiate an Analyzer object for given repository. It
// will return an error if the repository is not an actual folder.
func NewAnalyzer(repo string) (*Analyzer, error) {
	if !isDir(repo) {
		return nil, fmt.Errorf("provided repo '%s' is not a folder", repo)
	}
	return &Analyzer{
		stats: Stats{},
		repo:  repo,
	}, nil
}

// Analyze will walk through the source code and collect the dependency metrics.
func (a *Analyzer) Analyze() {
	path := a.repo
	if isDir(path) {
		a.analyzeDir(path)
	} else {
		a.analyzeFile(path)
	}
	return
}

// GetStats will return the collected dependency statistics.
func (a *Analyzer) GetStats() Stats {
	return a.stats
}

// analyzeDir will walk through all .go files in the folder and analyze each
// file using the analyzeFile method. It will skip the vendor folder if
// present.
func (a *Analyzer) analyzeDir(dirname string) {
	dirname = dirname + string(os.PathSeparator)
	filepath.Walk(dirname, func(path string, info os.FileInfo, err error) error {
		if err == nil && !info.IsDir() && !a.isVendor(path) && strings.HasSuffix(path, ".go") {
			a.analyzeFile(path)
		}
		return err
	})
	return
}

// analyzeFile will process the provided .go source file, and tracks its'
// external dependecies of packages within the same repo.
func (a *Analyzer) analyzeFile(fname string) {
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, fname, nil, 0)
	if err != nil {
		log.Printf("failed parsing %s: %s", fname, err)
	}
	for _, imp := range f.Imports {
		fo := getPackage(fname)
		fi := strings.Replace(imp.Path.Value, "\"", "", -1)
		st := a.stats[fo]
		if !a.inPackage(fi) {
			a.stats[fo] = st
			continue
		}
		st.FanOut++
		a.stats[fo] = st
		st = a.stats[fi]
		st.FanIn++
		a.stats[fi] = st
	}
}

// inPackage will check if the provided import is a local package or not.
func (a *Analyzer) inPackage(imp string) bool {
	return strings.Contains(imp, getPackage(a.repo))
}

// isVendor will check if the given path is the go vendor folder.
func (a *Analyzer) isVendor(path string) bool {
	return strings.Contains(
		filepath.ToSlash(filepath.Clean(path)),
		filepath.ToSlash(filepath.Clean(filepath.Join(a.repo, "vendor"))),
	)
}

// isDir will check if given filename is a directory or not.
func isDir(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && fi.IsDir()
}

// getPackage will return the fully qualified package name for given .go file.
func getPackage(fname string) string {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}
	file := filepath.ToSlash(filepath.Clean(filepath.Join(cwd, ".", fname)))
	root := filepath.ToSlash(filepath.Clean(filepath.Join(os.Getenv("GOPATH"), "src")))
	dir := filepath.ToSlash(filepath.Dir(strings.Replace(file, root+"/", "", -1)))
	return dir
}
