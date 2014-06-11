package main

import "flag"
import "fmt"
import "io/ioutil"
import "os"
import "path/filepath"
import "reporter"
import "strings"

const (
	help_text string = `
    Usage: goblin [command] [arguments]
    
    A lightweight static site generator written in Go.

          --help     display this help and exit
          --legal    displays legal notice and exit
          --version  output version information and exit

          -v, --verbose  show the names of the files in build
    `
	version_text = `0.1`
	legal_text   = `    
    Copyright (C) 2014 Abram C. Isola.
    This program comes with ABSOLUTELY NO WARRANTY; for details see
    LICENSE. This is free software, and you are welcome to redistribute 
    it under certain conditions in LICENSE.
`
)

var LOG *reporter.Reporter = reporter.NewReporter("goblin", nil)

func main() {
	help := flag.Bool("help", false, help_text)
	legal := flag.Bool("legal", false, legal_text)
	version := flag.Bool("version", false, version_text)
	verbose := flag.Bool("v", false, "be verbose")
	verbose2 := flag.Bool("verbose", false, "show the names of the files in build")
	flag.Parse()

	if *help {
		fmt.Println(help_text)
		os.Exit(0)
	}

	if *version {
		fmt.Println(version_text)
		os.Exit(0)
	}

	if *legal {
		fmt.Println(legal_text)
		os.Exit(0)
	}

	if flag.NArg() < 1 {
		LOG.Fatal("missing operands")
		os.Exit(1)
	}

	commands := flag.Args()
	if commands[0] == "init" {

		if len(commands) < 2 {
			LOG.Fatal("init requires another operand")
		}

		if Exists(commands[1]) {
			LOG.Fatal("cannot create site in an existing location")
		}

		os.Mkdir(commands[1], os.FileMode(0700))
		os.MkdirAll(filepath.Join(commands[1], "build"), os.FileMode(0700))
		os.MkdirAll(filepath.Join(commands[1], "src", "pages"), os.FileMode(0700))
		os.MkdirAll(filepath.Join(commands[1], "src", "posts"), os.FileMode(0700))
		os.MkdirAll(filepath.Join(commands[1], "themes", "default"), os.FileMode(0700))

		file, err := os.OpenFile(filepath.Join(commands[1], "config.json"), os.O_CREATE, 0600)
		if err != nil {
			LOG.Warningf("could not open create config.json: %s\n", err)
		}
		file.Close()

	} else if commands[0] == "build" {

		if !Exists("config.json") {
			LOG.Fatal("no config.json file found in this directory")
		}

		pagefiles, err := ioutil.ReadDir(filepath.Join(".", "src", "pages"))
		LOG.FatalOnError(err, "could not read directory '%s': %s", filepath.Join(".", "src", "pages"), err)

		// ensure they have a .md ending
		var pages []string
		for i := 0; i < len(pagefiles); i++ {
			if pagefiles[i].Name()[len(pagefiles[i].Name())-3:] == ".md" {
				pages = append(pages, pagefiles[i].Name())
			}
		}

		for i := 0; i < len(pages); i++ {
			fp, err := os.Open(filepath.Join(".", "src", "pages", pages[i]))
			LOG.FatalOnError(err, "could not open file '%s': %s", pages[i], err)

			contents, err := ioutil.ReadAll(fp)
			LOG.FatalOnError(err, "could not read file '%s': %s", pages[i], err)
			fp.Close()

			fp, err = os.OpenFile(filepath.Join(".", "build", strings.Replace(pages[i], ".md", ".html", -1)), os.O_CREATE, 0600)
			LOG.FatalOnError(err, "could not open file '%s': %s", strings.Replace(pages[i], ".md", ".html", -1), err)
			fp.Write(RenderMarkdown(contents))
			fp.Close()
		}

	} else if commands[0] == "serve" {
		manager := LoadConfig("config.json")
		manager.LoadPages()
		pages := manager.CheckPages()
		for i := 0; i < len(pages); i++ {
			if *verbose || *verbose2 { LOG.Infof("building %s", pages[i].Name) }
			fp, err := os.OpenFile(filepath.Join(".", "build", strings.Replace(pages[i].Name, ".md", ".html", -1)), os.O_CREATE, 0600)
			LOG.FatalOnError(err, "could not open file '%s': %s", strings.Replace(pages[i].Name, ".md", ".html", -1), err)
			fp.Write(RenderMarkdown(pages[i].Content))
			fp.Close()
		}
		manager.SaveRecords()
		// LOG.Fatal("feature 'serve' not yet implemented")
	} else if commands[0] == "deploy" {
		LOG.Fatal("feature 'deploy' not yet implemented")
	} else {
		LOG.Fatalf("invalid operation '%s'\n", commands[0])
	}
}

// Check if File / Directory Exists
func Exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		return false
	}
	return true
}
