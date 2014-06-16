package main

import "io/ioutil"
import "os"
import "path/filepath"
import "strconv"
import "strings"

type Page struct {
	Fi      os.FileInfo
	Raw     []byte
	Content string

	Title   string
	Author  string
	Layout  string
	Slug    string
	Url     string
	Mainnav bool
	Order   int
}

type Manager struct {
	Config      *Config
	Fspath      string
	Pages       []os.FileInfo
}

func LoadManager(path string) *Manager {
    config := LoadConfig(path)
    fspath, _ := filepath.Abs(filepath.Dir(path))
    man := &Manager{Config: config, Fspath: fspath}
    return man
}

func (m *Manager) LoadPages() {
	pagefiles, err := ioutil.ReadDir(filepath.Join(m.Fspath, "src", "pages"))
	OUT.FatalOnError(err, "could not read directory '%s': %s", filepath.Join(m.Fspath, "src", "pages"), err)

	for i := 0; i < len(pagefiles); i++ {
		m.Pages = append(m.Pages, pagefiles[i])
	}
}

func (m *Manager) SaveRecords() {
    // setup config
    config := NewConfig(filepath.Join(m.Fspath, ".goblinpages"))
    for i := 0; i < len(m.Pages); i++ {
		config.Set(m.Pages[i].Name(), m.Pages[i].ModTime().String())
	}
    SaveConfig(config)
}

func (m *Manager) CheckPages(all bool) []os.FileInfo {
    if all == false && Exists(filepath.Join(m.Fspath, ".goblinpages")) {
        gobpages := LoadConfig(filepath.Join(m.Fspath, ".goblinpages"))
        
        rpages := make([]os.FileInfo, 0)
        
        for i := 0; i < len(m.Pages); i++ {
            if gobpages.GetString(m.Pages[i].Name()) != m.Pages[i].ModTime().String() {
                rpages = append(rpages, m.Pages[i])
            }
            
        }
        return rpages
    }
	return m.Pages
}

func (m *Manager) loadpagevalues(page *Page) {
	unixraw := strings.Replace(string(page.Raw), "\r\n", "\n", -1)
	lines := strings.Split(unixraw, "\n")

	var found = 0
	for i, line := range lines {
		line = strings.TrimSpace(line)

		if found == 1 {
			// parse line for param
			colonIndex := strings.Index(line, ":")
			if colonIndex > 0 {
				key := strings.TrimSpace(line[:colonIndex])
				value := strings.TrimSpace(line[colonIndex+1:])
				value = strings.Trim(value, `"`) //remove quotes

				switch key {
				case "title":
					page.Title = value
				case "author":
					page.Author = value
				case "layout":
					page.Layout = value
				case "mainnav":
					if value == "true" {
						page.Mainnav = true
					} else {
						page.Mainnav = false
					}
				case "order":
					val, err := strconv.ParseInt(value, 10, 32)
					OUT.FatalOnError(err, "value of 'order' must be an integer in '%s'", page.Fi.Name())
					page.Order = int(val)
				case "url":
					page.Url = value
				case "slug":
					page.Slug = value
				}

			}

		} else if found >= 2 {
			// params over
			lines = lines[i:]
			break
		}

		if line == "---" {
			found += 1
		}

	}
	page.Content = strings.Join(lines, "\n")
}

func (m *Manager) LoadPage(fi os.FileInfo) Page {
	file, err := os.Open(filepath.Join(m.Fspath, "src", "pages", fi.Name()))
	OUT.FatalOnError(err, "could load '%s': %s", fi.Name(), err)

	page := Page{}
	page.Fi = fi
	page.Raw, err = ioutil.ReadAll(file)
	OUT.FatalOnError(err, "could load '%s': %s", fi.Name(), err)
	m.loadpagevalues(&page)

	file.Close()
	return page
}
