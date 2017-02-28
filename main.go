package main

import (
	"flag"
	"os"
	"strings"
)

var (
	modeFlags = map[string]*flag.FlagSet{
		"import":     flag.NewFlagSet("import", flag.PanicOnError),
		"search":     flag.NewFlagSet("search", flag.PanicOnError),
		"fetch_tags": flag.NewFlagSet("fetch_tags", flag.PanicOnError),
		"print":      flag.NewFlagSet("print", flag.PanicOnError),
	}
	modeTooltips = [][3]string{
		{
			"import",
			"PATHS...",
			"recursively imports all file and directory paths",
		},
		{
			"search",
			"TAGS..",
			"return paths to files that match a set of tags",
		},
		{
			"fetch_tags",
			"",
			"fetch tags for imported images and webm from gelbooru.com",
		},
		{
			"print",
			"",
			"prints the contents of the database for debuging purposes",
		},
	}
	deleteImported = modeFlags["import"].Bool(
		"d",
		false,
		"delete imported files",
	)
	addTagsToImported = modeFlags["import"].String(
		"t",
		"",
		"add tags to all imported files",
	)
	fetchTagsForImports = modeFlags["import"].Bool(
		"F",
		false,
		"fetch tags from gelbooru.com for imported files",
	)
)

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
		err = importPaths(
			fl.Args(),
			*deleteImported,
			*fetchTagsForImports,
			*addTagsToImported,
		)
	case "fetch_tags":
		err = fetchAllTags()
	case "print":
		err = printDB()
	case "search":
		err = searchPathsByTags(strings.Join(fl.Args(), " "))
	default:
		printHelp()
	}
	if err != nil {
		stderr.Println(err)
		os.Exit(1)
	}
}

func printHelp() {
	stderr.Println("Usage: hydron COMMAND [FLAGS...] ARGS...")
	for _, tt := range modeTooltips {
		stderr.Printf("\nhydron %s %s\n  %s\n", tt[0], tt[1], tt[2])
		modeFlags[tt[0]].PrintDefaults()
	}
	os.Exit(1)
}
