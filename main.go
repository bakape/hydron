package main

import (
	"flag"
	"log"
	"os"
)

var (
	stderr       = log.New(os.Stderr, "", 0)
	modeFlags    = map[string]*flag.FlagSet{}
	modeTooltips = map[string][2]string{
		"import": {
			"PATHS...",
			"recursively imports all file and directory paths",
		},
	}
)

func init() {
	modeFlags["import"] = flag.NewFlagSet("import", flag.PanicOnError)
}

func main() {
	if len(os.Args) < 2 {
		printHelp()
	}
	mode := os.Args[1]
	fl, ok := modeFlags[mode]
	if !ok {
		printHelp()
	}
	if err := fl.Parse(os.Args[2:]); err != nil {
		stderr.Println(err)
		printHelp()
	}
	if err := initDirs(); err != nil {
		panic(err)
	}
	if err := openDB(); err != nil {
		panic(err)
	}
	defer db.Close()

	var err error
	switch mode {
	case "import":
		err = importPaths(fl.Args())
	}
	if err != nil {
		stderr.Println(err)
		os.Exit(1)
	}
}

func printHelp() {
	stderr.Println("Usage: hydron COMMAND [FLAGS...] ARGS...")
	for _, c := range []string{"import"} {
		tt := modeTooltips[c]
		stderr.Printf("\nhydron %s %s\n\t%s\n", c, tt[0], tt[1])
		modeFlags[c].PrintDefaults()
	}
	os.Exit(1)
}
