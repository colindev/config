package config

import (
	"crypto/sha1"
	"fmt"
	"io/ioutil"
	"log"
	"time"
)

type Config struct {
	stopChan chan struct{}
	file     string
	hash     string
	parser   func(b []byte) (interface{}, error)
	Config   interface{}
}

func New(file string, parser func(b []byte) (interface{}, error)) (conf *Config, err error) {

	conf = &Config{
		stopChan: make(chan struct{}, 1),
		file:     file,
		parser:   parser,
	}

	b, hash, err := read(file)

	if err == nil {
		conf.hash = hash
		conf.Config, err = conf.parser(b)
	}

	return

}

func (conf *Config) Watch(sleepSuccess, sleepError time.Duration, fn func(interface{})) {

	for {
		select {
		case <-conf.stopChan:
		default:

			b, hash, err := read(conf.file)

			if err != nil {
				log.Println("read config error:", err)
				time.Sleep(sleepError * time.Second)
				continue
			}

			if conf.hash != hash {
				conf.hash = hash
				conf.Config, err = conf.parser(b)
				if err != nil {
					log.Println("config parse error:", err)
					time.Sleep(sleepError * time.Second)
					continue
				}

				fn(conf.Config)
			}

			time.Sleep(sleepSuccess * time.Second)
		}
	}
}

func (conf *Config) Stop() {
	conf.stopChan <- struct{}{}
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
