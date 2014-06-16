package main

import "fmt"
import "net/http"
import "os"
import "path/filepath"
import "strings"

import "github.com/aisola/reporter"
import "github.com/codegangsta/cli"
import "github.com/flosch/pongo"

const VERSION = "0.2"

var OUT = reporter.NewReporter("goblin", os.Stdout)

func main() {
    app := cli.NewApp()
    app.Name = "goblin"
    app.Usage = "a no-nonsense static site generator"
    app.Version = VERSION
    app.Flags = []cli.Flag{
        cli.BoolFlag{"verbose", "explain what you are doing"},
    }
    
    app.Commands = []cli.Command{
        
        {
            Name: "init",
            Usage: "initialize the static site directory",
            Action: func (ctx *cli.Context) {
                var site_directory string
                var argc = len(ctx.Args())
                
                // set where the site workspace will be
                if argc == 0 {
                    site_directory = "."
                } else if argc == 1 {
                    site_directory = ctx.Args().First()
                } else {
                    OUT.Fatal("init takes either zero or one value")
                }
                
                // setup site workspace
                if !Exists(site_directory) {
                    IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "creating '%s'", site_directory)
                    os.Mkdir(site_directory, os.FileMode(0700))  // it's a tough world, don't let others access
                }
                
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "creating '%s'", filepath.Join(site_directory, "build"))
                os.MkdirAll(filepath.Join(site_directory, "build"), os.FileMode(0700))
                
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "creating '%s'", filepath.Join(site_directory, "src", "pages"))
                os.MkdirAll(filepath.Join(site_directory, "src", "pages"), os.FileMode(0700))
                
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "creating '%s'", filepath.Join(site_directory, "src", "posts"))
                os.MkdirAll(filepath.Join(site_directory, "src", "posts"), os.FileMode(0700))
                
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "creating '%s'", filepath.Join(site_directory, "themes", "default"))
                os.MkdirAll(filepath.Join(site_directory, "themes", "default"), os.FileMode(0700))
                
                // setup config
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "creating '%s'", filepath.Join(site_directory, "config.json"))
                config := NewConfig(filepath.Join(site_directory, "config.json"))
                config.Set("url", "")
                config.Set("author", "")
                config.Set("theme", "default")
                SaveConfig(config)
                
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "creating '%s'", filepath.Join(site_directory, "src", "pages", "index.md"))
                err := CreateSimpleFile(filepath.Join(site_directory, "src", "pages", "index.md"),
                                        "\n---\ntitle: Home\nauthor: You\n\nlayout: page\nmainnav: true\norder: 0\nurl: /\nslug: home\n---\n\n##Home\n\nThis is the home page...\n\n",
                                        0644)
                if err != nil { OUT.Errorf("could not create index.md: %s", err) }
            },
        },
        
        {
            Name: "build",
            Usage: "build the static site",
            Flags: []cli.Flag{
                cli.BoolFlag{"all", "build all files regardless of the last modified date"},
            },
            Action: func (ctx *cli.Context) {
                var site_directory string
                var argc = len(ctx.Args())
                var html_name string
                
                // set where the site workspace will be
                if argc == 0 {
                    site_directory = "."
                } else if argc == 1 {
                    site_directory = ctx.Args().First()
                } else {
                    OUT.Fatal("serve takes either zero or one value")
                }
                
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "loading site configuration\n")
                manager := LoadManager(filepath.Join(site_directory, "config.json"))
                manager.LoadPages()
                pages := manager.CheckPages(ctx.IsSet("all"))
                
                for i := 0; i < len(pages); i++ {
                    IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "now building '%s'", pages[i].Name())
                    
                    page := manager.LoadPage(pages[i])
                    
                    if page.Url == "" {
                        html_name = filepath.Join(manager.Fspath, "build", strings.Replace(page.Fi.Name(), ".md", ".html", -1))
                    } else {
                        html_name = filepath.Join(manager.Fspath, "build", page.Url, strings.Replace(page.Fi.Name(), ".md", ".html", -1))
                    }
                    
                    layoutpath := filepath.Join(manager.Fspath, "themes", manager.Config.GetString("theme"), fmt.Sprintf("%s.html",page.Layout))
                    pongocontext := &pongo.Context{
                        "site_title": manager.Config.GetString("title"),
                        "site_url": manager.Config.GetString("url"),
                        "site_author": manager.Config.GetString("author"),
                        
                        "page_title": page.Title,
                        "page_author": page.Author,
                        
                        "content": string(RenderMarkdown(page.Content)),
                    }
                    
                    theme_out := RenderTheme(layoutpath, pongocontext)
                    
                    err := CreateSimpleFile(html_name, theme_out, 0644)
                    if err != nil { OUT.Errorf("could not build %s: %s", page.Fi.Name(), err) }
                }
                manager.SaveRecords()
                
            },
        },
        
        {
            Name: "serve",
            Usage: "run the static site in a server",
            Flags: []cli.Flag{
                cli.StringFlag{"port",":8080",`the server address to bind to (default: ":8080")`},
            },
            Action: func (ctx *cli.Context) {
                var site_directory string
                var argc = len(ctx.Args())
                
                // set where the site workspace will be
                if argc == 0 {
                    site_directory = "."
                } else if argc == 1 {
                    site_directory = ctx.Args().First()
                } else {
                    OUT.Fatal("serve takes either zero or one value")
                }
                
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "loading site configuration")
                manager := LoadManager(filepath.Join(site_directory, "config.json"))
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "serving '%s'\n", filepath.Join(manager.Fspath,"build"))
                err := http.ListenAndServe(":8080", http.FileServer(http.Dir(filepath.Join(manager.Fspath,"build"))))
                OUT.FatalOnError(err, "server had an error: %s", err)
            },
        },
        
    }
    
    app.Run(os.Args)
}
