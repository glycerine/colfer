package main

import (
	"flag"
	"io/ioutil"
	"log"
	"os"
	"path"
	"path/filepath"
	"strings"

	"github.com/pascaldekloe/colfer"
)

var (
	basedir = flag.String("b", ".", "Use a specific destination base `directory`.")
	prefix  = flag.String("p", "", "Adds a package `prefix`. Use slash as a separator when nesting.")
	format  = flag.Bool("f", false, "Normalizes schemas on the fly.")
	verbose = flag.Bool("v", false, "Enables verbose reporting to the standard error.")

	sizeMax = flag.String("s", "16 * 1024 * 1024", "Sets the default upper limit for serial byte sizes. The\n    \t`expression` is applied to the target language under the name\n    \tColferSizeMax.")
	listMax = flag.String("l", "64 * 1024", "Sets the default upper limit for the number of elements in a\n    \tlist. The `expression` is applied to the target language under the\n    \tname ColferListMax.")
)

var report = log.New(ioutil.Discard, "", 0)

func main() {
	flag.Parse()

	log.SetFlags(0)
	if *verbose {
		report.SetOutput(os.Stderr)
	}

	var files []string
	switch args := flag.Args(); len(args) {
	case 0:
		flag.Usage()
		os.Exit(2)
	case 1:
		files = []string{"."}
	default:
		files = args[1:]
	}

	// select language
	var gen func(string, []*colfer.Package) error
	switch lang := flag.Arg(0); strings.ToLower(lang) {
	case "c", "c++", "cpp":
		report.Println("Set up for C")
		log.Fatal("colf: C template not implemented yet")
	case "go":
		report.Println("Set up for Go")
		gen = colfer.GenerateGo
	case "java":
		report.Println("Set up for Java")
		gen = colfer.GenerateJava
	case "javascript", "js", "ecmascript":
		report.Println("Set up for ECMAScript")
		gen = colfer.GenerateECMA
	default:
		log.Fatalf("colf: unsupported language %q", lang)
	}

	// resolve clean file set
	var writeIndex int
	for i := 0; i < len(files); i++ {
		f := files[i]

		info, err := os.Stat(f)
		if err != nil {
			log.Fatal(err)
		}
		if info.IsDir() {
			colfFiles, err := filepath.Glob(filepath.Join(f, "*.colf"))
			if err != nil {
				log.Fatal(err)
			}
			files = append(files, colfFiles...)
			continue
		}

		f = filepath.Clean(f)
		for j := 0; ; j++ {
			if j == writeIndex {
				files[writeIndex] = f
				writeIndex++
				break
			}
			if files[j] == f {
				report.Println("Duplicate inclusion of", f, "ignored")
				break
			}
		}
	}
	files = files[:writeIndex]
	report.Println("Found schema files", strings.Join(files, ", "))

	packages, err := colfer.ParseFiles(files)
	if err != nil {
		log.Fatal(err)
	}

	if *format {
		for _, file := range files {
			changed, err := colfer.Format(file)
			if err != nil {
				log.Fatal(err)
			}
			if changed {
				log.Println("colfer: formatted", file)
			}
		}
	}

	if len(packages) == 0 {
		log.Fatal("colfer: no struct definitons found")
	}

	for _, p := range packages {
		p.Name = path.Join(*prefix, p.Name)
		p.SizeMax = *sizeMax
		p.ListMax = *listMax
	}

	if err := gen(*basedir, packages); err != nil {
		log.Fatal(err)
	}
}

// ANSI escape codes for markup
const (
	bold      = "\x1b[1m"
	underline = "\x1b[4m"
	clear     = "\x1b[0m"
)

func init() {
	cmd := os.Args[0]

	help := bold + "NAME\n\t" + cmd + clear + " \u2014 compile Colfer schemas\n\n"
	help += bold + "SYNOPSIS\n\t" + cmd + clear
	help += " [ " + underline + "options" + clear + " ] " + underline + "language" + clear
	help += " [ " + underline + "file" + clear + " " + underline + "..." + clear + " ]\n\n"
	help += bold + "DESCRIPTION\n\t" + clear
	help += "Generates source code for a " + underline + "language" + clear + ". The options are: "
	help += bold + "C" + clear + ", " + bold + "Go" + clear + ",\n"
	help += "\t" + bold + "Java" + clear + " and " + bold + "JavaScript" + clear + ".\n"
	help += "\tThe " + underline + "file" + clear + " operands specify the input. Directories are scanned for\n"
	help += "\tfiles with the colf extension. If " + underline + "file" + clear + " is absent, " + cmd + " includes\n"
	help += "\tthe working directory.\n"
	help += "\tA package can have multiple schema files.\n\n"
	help += bold + "OPTIONS\n" + clear

	tail := "\n" + bold + "EXIT STATUS" + clear + "\n"
	tail += "\tThe command exits 0 on succes, 1 on compilation failure and 2\n"
	tail += "\twhen invoked without arguments.\n"
	tail += "\n" + bold + "EXAMPLES" + clear + "\n"
	tail += "\tCompile ./api/*.colf into ./src/ as Java:\n\n"
	tail += "\t\t" + cmd + " -p com/example -b src java api\n\n"
	tail += "\tCompile ./io.colf with compact limits as C:\n\n"
	tail += "\t\t" + cmd + " -s 2048 -l 96 c io.colf\n"
	tail += "\n" + bold + "BUGS" + clear + "\n"
	tail += "\tReport bugs at https://github.com/pascaldekloe/colfer/issues\n\n"
	tail += bold + "SEE ALSO\n\t" + clear + "protoc(1)\n"

	flag.Usage = func() {
		os.Stderr.WriteString(help)
		flag.PrintDefaults()
		os.Stderr.WriteString(tail)
	}
}
