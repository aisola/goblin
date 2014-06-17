package main

import "fmt"
import "io/ioutil"
import "net/http"
import "os"
import "path/filepath"
import "strings"

import "github.com/aisola/reporter"
import "github.com/codegangsta/cli"
import "github.com/flosch/pongo"

const VERSION = "0.3"

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
            Usage: "initialize the workspace",
            Description: "The init command initializes the static site, putting it in the given \n   directory if supplied. If the given directory does not exist, it will \n   be created.",
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
        
        {   // TODO: impliment posts
            Name: "build",
            Usage: "build the static site",
            Description: "The build command compiles each of the pages and posts into html and \n   matches them with their layout. The build will only build files that \n   have not been modified since their last build. If the all/a option is \n   set all of the pages/posts will be compiled regardless of whether \n   they have have been modified or not.",
            Flags: []cli.Flag{
                cli.BoolFlag{"all, a", "build all files regardless of the last modified date"},
                // TODO: cli.BoolFlag{"pages, p", "pages build only"},
                // TODO: cli.BoolFlag{"posts", "build posts only"},
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
                    OUT.Fatal("build takes either zero or one value")
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
                        err := os.MkdirAll(filepath.Join(manager.Fspath, "build", page.Url), 0755)
                        OUT.FatalOnError(err, "cannot create necessary directory: %s", err)
                        html_name = filepath.Join(manager.Fspath, "build", page.Url, "index.html")
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
                
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "copying theme static directory\n")
                
                staticdir_theme := filepath.Join(manager.Fspath, "themes", manager.Config.GetString("theme"), "static")
                staticdir_build := filepath.Join(manager.Fspath, "build", "static")
                
                err := CopyDir(staticdir_theme, staticdir_build)
                OUT.FatalOnError(err, "cannot copy static directory: %s", err)
            },
        },
        
        {
            Name: "serve",
            Usage: "run the static site in a server",
            Description: "The serve command creates a server (rooted in the workspace 'build' \n   'build' directory). The bind option sets where the server should \n   serve. (default: localhost:8080)",
            Flags: []cli.Flag{
                cli.StringFlag{"bind",":8080",`the server address to bind to (default: ":8080")`},
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
        
        {
            Name: "clean-build",
            Usage: "remove all files from the build directory (clean start)",
            Description: "The clean-build command will remove all files under the \n   workspace build directory.",
            Action: func (ctx *cli.Context) {
                var site_directory string
                var argc = len(ctx.Args())
                
                // set where the site workspace will be
                if argc == 0 {
                    site_directory = "."
                } else if argc == 1 {
                    site_directory = ctx.Args().First()
                } else {
                    OUT.Fatal("clean-build takes either zero or one value")
                }
                
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "removing '%s'\n", filepath.Join(site_directory,".goblinpages"))
                err := os.Remove(filepath.Join(site_directory,".goblinpages"))
                if err != nil { OUT.Errorf("error removing '.goblinpages': %s", err) }
                
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "removing '%s'\n", filepath.Join(site_directory,".goblinposts"))
                err = os.Remove(filepath.Join(site_directory,".goblinposts"))
                if err != nil { OUT.Errorf("error removing '.goblinposts': %s", err) }
                
                IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "reading '%s'\n", filepath.Join(site_directory,"build"))
                files, err := ioutil.ReadDir(filepath.Join(site_directory, "build"))
                if err != nil { OUT.Errorf("error reading '%s': %s", site_directory, err) }
                
                for _, file := range files {
                    IfTrueExec(ctx.GlobalBool("verbose"), OUT.Infof, "removing '%s'\n", filepath.Join(site_directory,"build",file.Name()))
                    err := os.RemoveAll(filepath.Join(site_directory,"build",file.Name()))
                    if err != nil { OUT.Errorf("error removing: %s", err) }
                }
            },
        },
    }
    
    app.Run(os.Args)
}
