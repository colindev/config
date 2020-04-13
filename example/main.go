package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"time"

	config "github.com/colindev/config-watcher"
	yaml "gopkg.in/yaml.v2"
)

func main() {

	var (
		o struct {
			Name string `yaml:"name"`
			Num  int    `yaml:"num"`
		}
		fname = "./x"
	)

	f, err := os.OpenFile(fname, os.O_RDWR|os.O_CREATE, 0644)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	defer os.Remove(fname)

	f.WriteString(`
name: xxx
num: 0
`)

	go func() {
		n := 0
		for {
			time.Sleep(time.Second * 10)
			n = n + 1
			f.Truncate(0)
			f.Seek(0, 0)
			f.WriteString(fmt.Sprintf(`
name: xxx
num: %d
`, n))
		}
	}()

	conf, err := config.New(fname, func(b []byte) (interface{}, error) {
		err := yaml.Unmarshal(b, &o)

		return o, err
	})

	if err != nil {
		log.Fatal(err)
	}

	go func() {
		time.Sleep(60 * time.Second)
		conf.Stop()
	}()

	go func() {
		for {
			log.Println(conf.Config())
			time.Sleep(5 * time.Second)
		}
	}()

	go func() {

		buf := bufio.NewReader(os.Stdin)
		for {
			line, _, err := buf.ReadLine()
			if err != nil {
				log.Println("<error:", err)
				continue
			}
			log.Println("<", string(line))
			conf.UpdateNow()
		}

	}()

	conf.Watch(func(o interface{}) {
		log.Printf("watch: %#v\n", o)
	})

}
