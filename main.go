package main

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

var (
	modeFlags = map[string]*flag.FlagSet{
		"import": flag.NewFlagSet("import", flag.PanicOnError),
		"search": flag.NewFlagSet("search", flag.PanicOnError),
	}
	modeTooltips = [][3]string{
		{
			"import",
			"PATHS...",
			"Recursively import all file and directory PATHS.",
		},
		{
			"search",
			"TAGS...",
			"Return paths to files that match the set of TAGS." + `
  TAGS can include prefixed system tags for searching by image metadata:
    size, width, height, length, tag_count,
  followed by one of these comparison operators:
    >, <, =
  and a positive number.
  Examples:
    hydron search 'system:width>1920 system:height>1080'
    hydron search 'system:tag_count=0
    hydron search 'red_scarf system:size<10485760'`,
		},
		{
			"complete_tag",
			"PREFIX",
			"Suggest tags that start with PREFIX for autocompletion.",
		},
		{
			"add_tags",
			"ID TAGS...",
			"Add TAGS... to file specified by hex-encoded SHA1 hash ID.",
		},
		{
			"fetch_tags",
			"",
			"Fetch tags for imported images and webm from gelbooru.com.",
		},
		{
			"print",
			"",
			"Print the contents of the database for debuging purposes.",
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
	returnRandom = modeFlags["search"].Bool(
		"r",
		false,
		"return only one random matched file",
	)
)

func main() {
	assertArgCount(2)
	mode := os.Args[1]
	fl, ok := modeFlags[mode]
	if ok {
		if err := fl.Parse(os.Args[2:]); err != nil {
			stderr.Println(err)
			printHelp()
		}
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
		err = searchPathsByTags(strings.Join(fl.Args(), " "), *returnRandom)
	case "complete_tag":
		assertArgCount(3)
		tags := completeTag(os.Args[2])
		fmt.Println(strings.Join(tags, " "))
	case "add_tags":
		assertArgCount(4)
		err = addTagsCLI(os.Args[2], os.Args[3:])
	default:
		printHelp()
	}
	if err != nil {
		stderr.Println(err)
		os.Exit(1)
	} else {
		os.Exit(0)
	}
}

func assertArgCount(i int) {
	if len(os.Args) < i {
		printHelp()
	}
}

func printHelp() {
	stderr.Println("Usage: hydron COMMAND [FLAGS...] [ARGS...]")
	for _, tt := range modeTooltips {
		stderr.Printf("\nhydron %s %s\n  %s\n", tt[0], tt[1], tt[2])
		flags := modeFlags[tt[0]]
		if flags != nil {
			stderr.Print("\n")
			flags.PrintDefaults()
		}
	}
	os.Exit(1)
}
