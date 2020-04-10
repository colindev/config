package config

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"sync"

	"github.com/fsnotify/fsnotify"
)

func init() {
	log.SetFlags(log.Lshortfile)
}

type Config struct {
	stopChan   chan struct{}
	updateChan chan struct{}
	filePath   string
	hash       string
	parser     func(b []byte) (interface{}, error)
	config     interface{}
	sync.RWMutex
	wc *fsnotify.Watcher
}

func New(fname string, parser func(b []byte) (interface{}, error)) (conf *Config, err error) {

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, err
	}

	err = watcher.Add(fname)
	if err != nil {
		return nil, err
	}

	conf = &Config{
		stopChan: make(chan struct{}, 1),
		filePath: fname,
		parser:   parser,
		wc:       watcher,
	}

	err = conf.init()

	return

}

func (conf *Config) Config() interface{} {
	conf.RLock()
	c := conf.config
	conf.RUnlock()

	return c
}

func (conf *Config) Watch(updater func(interface{})) {

	for {
		select {
		case <-conf.stopChan:
			return
		case err, ok := <-conf.wc.Errors:
			if !ok {
				return
			}
			log.Println("fsnotify error:", err)
		case event, ok := <-conf.wc.Events:

			if !ok {
				return
			}

			log.Printf("fsnotify event: %#v %s\n", event, event)
			if err := conf.update(updater); err != nil {
				log.Println("update error:", err)
			}
		}
	}
}

func (conf *Config) init() error {

	b, hash, err := read(conf.filePath)
	if err != nil {
		return err
	}
	conf.Lock()
	conf.hash = hash
	conf.config, err = conf.parser(b)
	conf.Unlock()
	if err != nil {
		return err
	}

	return nil
}

func (conf *Config) update(updater func(interface{})) error {

	b, hash, err := read(conf.filePath)

	if err != nil {
		return err
	}

	if conf.hash != hash {
		conf.Lock()
		conf.hash = hash
		conf.config, err = conf.parser(b)
		conf.Unlock()
		if err != nil {
			return err
		}

		updater(conf.config)
	}
	return nil
}

func (conf *Config) Stop() {
	conf.stopChan <- struct{}{}
	conf.wc.Close()
}

func read(file string) (b []byte, hash string, err error) {

	b, err = ioutil.ReadFile(file)
	if err != nil {
		return
	}
	h := sha1.New()
	h.Write(b)
	hash = fmt.Sprintf("%x", h.Sum(nil))

	return
}
