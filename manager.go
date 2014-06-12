package main

import "encoding/json"
import "io/ioutil"
import "os"
import "path/filepath"
import "strconv"
import "strings"
import "time"

type Record struct {
	Name string
	Mod  time.Time
}

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
	Author      string
	Url         string
	Theme       string
	Fspath      string
	Pages       []os.FileInfo
	pagerecords []*Record
}

func LoadConfig(configpath string) *Manager {
	config, err := os.Open(configpath)
	LOG.FatalOnError(err, "could not open config file '%s': %s", configpath, err)
	configdecoder := json.NewDecoder(config)
	man := &Manager{}
	err = configdecoder.Decode(man)
	LOG.FatalOnError(err, "could not decode config file '%s': %s", configpath, err)
	man.Fspath = filepath.Dir(filepath.Clean(configpath))
	return man
}

func (m *Manager) SaveRecords() error {
	pagefiles, err := ioutil.ReadDir(filepath.Join(m.Fspath, "src", "pages"))
	if err != nil {
		return err
	}

	if Exists(filepath.Join(m.Fspath, ".goblinpages")) {
		os.Remove(filepath.Join(m.Fspath, ".goblinpages"))
	}

	pages := make([]*Record, 0)
	for i := 0; i < len(pagefiles); i++ {
		pages = append(pages, &Record{Name: pagefiles[i].Name(), Mod: pagefiles[i].ModTime()})
	}

	data, err := json.Marshal(pages)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(filepath.Join(m.Fspath, ".goblinpages"), os.O_CREATE, 0600)
	if err != nil {
		return err
	}

	file.Write(data)
	file.Close()
	return nil
}

func (m *Manager) LoadPages() {
	pagefiles, err := ioutil.ReadDir(filepath.Join(m.Fspath, "src", "pages"))
	LOG.FatalOnError(err, "could not read directory '%s': %s", filepath.Join(m.Fspath, "src", "pages"), err)

	for i := 0; i < len(pagefiles); i++ {
		m.Pages = append(m.Pages, pagefiles[i])
	}
}

func (m *Manager) CheckPages() []os.FileInfo {
	if Exists(filepath.Join(m.Fspath, ".goblinpages")) {
		file, err := os.Open(filepath.Join(m.Fspath, ".goblinpages"))
		LOG.FatalOnError(err, "couldn't open '.goblinpages': %s", err)
		rdecoder := json.NewDecoder(file)
		err = rdecoder.Decode(&m.pagerecords)
		LOG.FatalOnError(err, "couldn't decode '.goblinpages': %s", err)

		rpages := make([]os.FileInfo, 0)
		for i := 0; i < len(m.pagerecords); i++ {
			for i2 := 0; i2 < len(m.Pages); i2++ {
				if m.pagerecords[i].Name == m.Pages[i2].Name() {
					if m.pagerecords[i].Mod.String() != m.Pages[i2].ModTime().String() {
						rpages = append(rpages, m.Pages[i2])
					}
				}
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
					LOG.FatalOnError(err, "value of 'order' must be an integer in '%s'", page.Fi.Name())
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
	LOG.FatalOnError(err, "could load '%s': %s", fi.Name(), err)

	page := Page{}
	page.Fi = fi
	page.Raw, err = ioutil.ReadAll(file)
	LOG.FatalOnError(err, "could load '%s': %s", fi.Name(), err)
	m.loadpagevalues(&page)

	file.Close()
	return page
}
