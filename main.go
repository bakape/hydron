package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/bakape/hydron/db"
	"github.com/bakape/hydron/files"
	"github.com/bakape/hydron/util"
)

const defaultAddress = ":8010"

var (
	modeFlags = map[string]*flag.FlagSet{
		"serve":  flag.NewFlagSet("serve", flag.PanicOnError),
		"import": flag.NewFlagSet("import", flag.PanicOnError),
		"search": flag.NewFlagSet("search", flag.PanicOnError),
	}
	modeTooltips = [][3]string{
		{
			"serve",
			"(default)",
			`Launch hydron in server mode, exposing files and commands through
  a HTTP/JSON API until terminated.`,
		},
		{
			"import",
			"PATHS...",
			"Recursively import all file and directory PATHS.",
		},
		{
			"remove",
			"IDs...",
			"Remove files specified by hex-encoded SHA1 hash IDs.",
		},
		{
			"search",
			"TAGS...",
			"Return paths to files that match the set of TAGS." + `
  TAGS can be prefixed with - to match a subset that does not include this tag.
  TAGS can be prefixed to match a specific tag category like artist, series and
  character.
  TAGS can include an order:$x parameter where $x is one of:
	  size, width, height, duration, tag_count, random.
  Prefixing - before $x will reverse the order.
  TAGS can include prefixed system tags for searching by file metadata:
    size, width, height, duration, tag_count,
  followed by one of these comparison operators:
    >, <, =, >=, <=
  and a positive integer.
  Examples:
    hydron search system:width>1920 system:height>1080 artist:null
    hydron search system:tag_count=0 order:random
    hydron search 'red_scarf -bed system:size<10485760'`,
		},
		{
			"complete_tag",
			"PREFIX",
			"Suggest tags that start with PREFIX for autocompletion.",
		},
		{
			"add_tags",
			"ID TAGS...",
			"Add TAGS to file specified by hex-encoded SHA1 hash ID.",
		},
		{
			"remove_tags",
			"ID TAGS...",
			"Remove TAGS from file specified by hex-encoded SHA1 hash ID.",
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
		"f",
		false,
		"Fetch tags from gelbooru.com for imported files.\n"+
			"NB: This will notably slow down importing large amounts of files.\n"+
			"Consider using import, followed by fetch_tags!",
	)
	address = modeFlags["serve"].String(
		"a",
		defaultAddress,
		"address to listen on for requests",
	)
	page = modeFlags["search"].Uint64(
		"p",
		0,
		"page of the query to view",
	)
)

func main() {
	var (
		mode string
		fl   *flag.FlagSet
	)
	if len(os.Args) == 1 {
		mode = "serve"
		*address = defaultAddress
	} else {
		assertArgCount(2)
		mode = os.Args[1]

		var ok bool
		fl, ok = modeFlags[mode]
		if ok {
			if err := fl.Parse(os.Args[2:]); err != nil {
				stderr.Println(err)
				printHelp()
			}
		}
	}

	if err := util.Waterfall(files.Init, db.Open); err != nil {
		panic(err)
	}
	defer db.Close()

	var err error
	switch mode {
	case "serve":
		err = startServer(*address)
	case "import":
		assertArgCount(3)
		err = importPaths(
			fl.Args(),
			*deleteImported,
			*fetchTagsForImports,
			*addTagsToImported,
		)
	case "remove":
		assertArgCount(3)
		err = removeFiles(os.Args[2:])
	case "fetch_tags":
		err = fetchAllTags()
	case "search":
		err = searchImages(strings.Join(fl.Args(), " "), int(*page))
	case "complete_tag":
		assertArgCount(3)
		var suggests []string
		suggests, err = db.CompleTag(os.Args[2])
		fmt.Println(strings.Join(suggests, " "))
	case "add_tags":
		assertArgCount(4)
		err = addTags(os.Args[2], strings.Join(os.Args[3:], " "))
	case "remove_tags":
		assertArgCount(4)
		err = removeTags(os.Args[2], strings.Join(os.Args[3:], " "))
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
