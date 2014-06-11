package main

import "encoding/json"
import "io/ioutil"
import "os"
import "path/filepath"
import "time"

type Record struct {
	Name string
	Mod  time.Time
}

type Page struct {
	Title  string
	Name   string
	Slug   string
	Url    string
	Author string
	Date   string
	Layout string

	Content []byte
	Lastmod time.Time
}

type Manager struct {
	Author  string
	Url     string
	Theme   string
	Fspath  string
	Pages   []*Page
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
	for i:=0; i<len(m.Pages); i++ {
		m.pagerecords = append(m.pagerecords, &Record{Name: m.Pages[i].Name, Mod: m.Pages[i].Lastmod})
	}

	data, err := json.Marshal(m.pagerecords)
	if err != nil { return err }

	file, err := os.OpenFile(filepath.Join(m.Fspath, ".goblinpages"), os.O_CREATE, 0600)
	if err != nil { return err }

	file.Write(data)
	file.Close()

	return nil
}

func (m *Manager) LoadPages() {
	pagefiles, err := ioutil.ReadDir(filepath.Join(m.Fspath, "src", "pages"))
	LOG.FatalOnError(err, "could not read directory '%s': %s", filepath.Join(m.Fspath, "src", "pages"), err)

	// ensure they have a .md ending
	var pages []string
	for i := 0; i < len(pagefiles); i++ {
		if pagefiles[i].Name()[len(pagefiles[i].Name())-3:] == ".md" {
			pages = append(pages, pagefiles[i].Name())
		}
	}

	for i := 0; i < len(pages); i++ {
		fp, err := os.Open(filepath.Join(m.Fspath, "src", "pages", pages[i]))
		LOG.FatalOnError(err, "could not open file '%s': %s", pages[i], err)

		contents, err := ioutil.ReadAll(fp)
		LOG.FatalOnError(err, "could not read file '%s': %s", pages[i], err)
		fp.Close()

		m.Pages = append(m.Pages, &Page{Content:contents, Name: pages[i]})
	}
}

func (m *Manager) CheckPages() []*Page {
	if Exists(filepath.Join(m.Fspath, ".goblinpages")) {
		file, err := os.Open(filepath.Join(m.Fspath, ".goblinpages"))
		LOG.FatalOnError(err, "couldn't open '.goblinpages': %s", err)
		rdecoder := json.NewDecoder(file)
		err = rdecoder.Decode(&m.pagerecords)
		LOG.FatalOnError(err, "couldn't decode '.goblinpages': %s", err)

		rpages := make([]*Page, 0)
		for i:=0; i<len(m.pagerecords); i++ {
			for i2:=0; i2<len(m.Pages); i2++ {
				if m.pagerecords[i].Name == m.Pages[i2].Name {
					if m.pagerecords[i].Mod.String() != m.Pages[i2].Lastmod.String() {
						rpages = append(rpages, m.Pages[i2])
					}
				}
			}
		}
		return rpages
	}
	return m.Pages
}

// func (m *Manager) LoadPage() *Page {}


// func (m *Manager) CheckPosts() {}
// func (m *Manager) LoadPosts() {}
// func (m *Manager) LoadPost() {}
