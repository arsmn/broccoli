package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"
)

var (
	flagInput    = flag.String("src", "public", "")
	flagOutput   = flag.String("o", "", "")
	flagVariable = flag.String("var", "br", "")
	flagInclude  = flag.String("include", "", "")
	flagExclude  = flag.String("exclude", "", "")
	flagQuality  = flag.Int("quality", 11, "")

	verbose = flag.Bool("v", false, "")
)

const (
	constInput = "public"
)

const help = `Usage: broccoli [options]

Broccoli uses brotli compression to embed a virtual file system in Go executables.

Options:
	-src folder[,file,file2]
		The input files and directories, "public" by default.
	-o
		Name of the generated file, follows input by default.
	-var=br
		Name of the exposed variable, "br" by default.
	-include *.html,*.css
		Wildcard for the files to include, no default.
	-include *.wasm
		Wildcard for the files to include, no default.
	-quality [level]
		Brotli compression level (0-11), the highest by default.

Generate a broccoli.gen.go file with the variable broccoli:
	//go:generate broccoli -src assets -o broccoli -var broccoli

Generate a regular public.gen.go file, but include all *.wasm files:
	//go:generate broccoli -src public -include="*.wasm"`

var goIdentifier = regexp.MustCompile(`^\p{L}[\p{L}0-9_]*$`)

func main() {
	log.SetFlags(0)
	log.SetPrefix("broccoli: ")
	flag.Usage = func() {
		fmt.Fprintln(os.Stderr, help)
	}

	flag.Parse()
	if len(os.Args) == 1 {
		flag.Usage()
	}

	var inputs []string
	if flagInput == nil {
		inputs = []string{constInput}
	} else {
		inputs = strings.Split(*flagInput, ",")
	}

	output := *flagOutput
	if output == "" {
		output = inputs[0]
	}
	if !strings.HasSuffix(output, ".gen.go") {
		output = strings.Split(output, ".")[0] + ".gen.go"
	}

	variable := *flagVariable
	if !goIdentifier.MatchString(variable) {
		log.Fatalln(variable, "is not a valid Go identifier")
	}

	includeGlob := *flagInclude
	excludeGlob := *flagExclude
	if includeGlob != "" && excludeGlob != "" {
		log.Fatal("mutually exclusive options -include and -include found")
	}

	quality := *flagQuality
	if quality < 1 || quality > 11 {
		log.Fatalf("unsupported compression level %d (1-11)\n", quality)
	}

	g := Generator{
		inputFiles:  inputs,
		outputFile:  output,
		outputVar:   variable,
		includeGlob: includeGlob,
		excludeGlob: excludeGlob,
		quality:     quality,
	}

	g.parsePackage()
	g.generate()
}
