package icu

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/BurntSushi/toml"
)

type Bundle interface {
	TranslatorFor(tag Tag) Translator
}

var (
	_ Bundle = (*DirectoryBundle)(nil)
)

func NewDirectoryBundle(directory string) *DirectoryBundle {
	return &DirectoryBundle{
		directory: directory,
		cache:     map[Tag]cacheEntry{},
	}
}

type DirectoryBundle struct {
	directory string
	mu        sync.RWMutex
	cache     map[Tag]cacheEntry
}

func (b *DirectoryBundle) TranslatorFor(tag Tag) Translator {
	if b == nil {
		return nilTranslator
	}
	t, err := b.load(tag)
	if err != nil {
		return nilTranslator
	}
	return t
}

func (b *DirectoryBundle) load(tag Tag) (*HierachicalTranslator, error) {
	f := filepath.Join(b.directory, fmt.Sprintf("%s.toml", tag))
	var err error

	fi, err := os.Stat(f)
	if err != nil {
		return nil, err
	}

	// Do we have a cached version?
	b.mu.RLock()
	e, ok := b.cache[tag]
	b.mu.RUnlock()
	if ok && fi.ModTime().Equal(e.modTime) {
		return e.translator, nil
	}

	// No cached version or old version. Load language from file.
	t := NewHierachicalTranslator()
	_, err = toml.DecodeFile(f, t)
	if err != nil {
		return nil, err
	}

	// Cache loaded version.
	b.mu.Lock()
	b.cache[tag] = cacheEntry{translator: t, modTime: fi.ModTime()}
	b.mu.Unlock()

	return t, nil
}

type cacheEntry struct {
	translator *HierachicalTranslator
	modTime    time.Time
}
