package main

import (
	"encoding/json"
	"os"
	"sort"
)

// A Pinlist represents a pinned list of dependencies.
type Pinlist struct {
	Deps     []Dep `json:"deps"`
}

// A Dep is a third-party package dependency.
type Dep struct {
	// Name is the import path of the package.
	Name string `json:"name"`
	// Rev is the VCS revision number (e.g., git SHA-1 hash).
	Rev string `json:"rev"`
}

func LoadPinlist(filename string) (*Pinlist, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	decoder := json.NewDecoder(f)
	var p Pinlist
	if err := decoder.Decode(&p); err != nil {
		return nil, err
	}
	return &p, nil
}

func (p *Pinlist) Save(filename string) error {
	p.normalize()
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	b, err := json.MarshalIndent(p, "", "\t")
	if err != nil {
		return err
	}
	_, err = f.Write(b)
	return err
}

type deps []Dep

func (d deps) Len() int           { return len(d) }
func (d deps) Less(i, j int) bool { return d[i].Name < d[j].Name }
func (d deps) Swap(i, j int)      { d[i], d[j] = d[j], d[i] }

func (p *Pinlist) normalize() {
	sort.Sort(deps(p.Deps))
}
