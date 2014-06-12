package main

import "flag"
import "fmt"
import "github.com/aisola/reporter"
import "os"
import "path/filepath"
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

		manager := LoadConfig("config.json")
		manager.LoadPages()
		pages := manager.CheckPages()

		for i := 0; i < len(pages); i++ {
			if *verbose || *verbose2 {
				LOG.Infof("building %s", pages[i].Name())
			}

			page := manager.LoadPage(pages[i])

			html_name := filepath.Join(manager.Fspath, "build", strings.Replace(page.Fi.Name(), ".md", ".html", -1))

			file, err := os.OpenFile(html_name, os.O_CREATE, 0644)
			LOG.FatalOnError(err, "could not open file '%s': %s", html_name, err)

			file.Write(RenderMarkdown([]byte(page.Content)))
			file.Close()
		}
		manager.SaveRecords()

	} else if commands[0] == "serve" {
		LOG.Fatal("feature 'serve' not yet implemented")
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
