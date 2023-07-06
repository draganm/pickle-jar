package jsfiles

import (
	"os"
	"path/filepath"
	"strings"
)

type JSFile struct {
	Path    string
	Content string
}

type Dir struct {
	files    []JSFile
	parent   *Dir
	children map[string]*Dir
}

func New() *Dir {
	return &Dir{
		files:    []JSFile{},
		parent:   nil,
		children: map[string]*Dir{},
	}
}

func (d *Dir) AddFile(path, content string) {

	vol := filepath.VolumeName(path)

	withoutVolume := filepath.Clean(path[len(vol):])
	d.addFile(withoutVolume, content)
}

func (d *Dir) AllStepDefinitions(path string) []JSFile {
	vol := filepath.VolumeName(path)

	withoutVolume := filepath.Clean(path[len(vol):])

	return d.allStepDefinitions(withoutVolume)
}

func (d *Dir) allStepDefinitions(path string) []JSFile {
	allFiles := []JSFile{}

	sdChild, found := d.children["step_definitions"]
	if found {
		for _, f := range sdChild.files {
			allFiles = append(allFiles, JSFile{Path: filepath.Join("step_definitions", f.Path), Content: f.Content})
		}
	}

	before, after, found := strings.Cut(path, string(os.PathSeparator))

	if !found || before == "" {
		return allFiles
	}

	ch, found := d.children[before]
	if found {
		childFiles := ch.allStepDefinitions(after)
		for _, f := range childFiles {
			allFiles = append(allFiles, JSFile{Path: filepath.Join(before, f.Path), Content: f.Content})
		}
	}

	return allFiles

}

func (d *Dir) addFile(path, content string) {
	before, after, found := strings.Cut(path, string(os.PathSeparator))

	if !found || before == "" {
		d.files = append(d.files, JSFile{Path: path, Content: content})
		return
	}

	ch, found := d.children[before]
	if !found {
		ch = &Dir{
			files:    []JSFile{},
			parent:   d,
			children: map[string]*Dir{},
		}
		d.children[before] = ch
	}

	ch.addFile(after, content)

}
