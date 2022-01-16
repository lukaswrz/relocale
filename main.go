package main

import (
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/jessevdk/go-flags"
	"golang.org/x/text/language"

	"github.com/lukaswrz/relocale/config"
)

type options struct {
	Config string `short:"c" long:"config" description:"Configuration file" value-name:"FILE"`
}

func main() {
	opts := options{}

	args, err := flags.Parse(&opts)
	if err != nil {
		os.Exit(1)
	}

	if len(args) > 0 {
		log.Fatal(errors.New("Unexpected operand(s)"))
	}

	var cf string
	if opts.Config == "" {
		res := config.Locate()

		if res == "" {
			log.Fatal(errors.New("Configuration file not found"))
		}

		cf = res
	} else {
		res := opts.Config

		_, err := os.Stat(res)
		if err != nil {
			log.Fatalf("Specified configuration file not accessible: %s", err.Error())
		}

		cf = res
	}

	content, err := ioutil.ReadFile(cf)
	if err != nil {
		log.Fatalf("Unable to read file: %s", err.Error())
	}

	c, err := config.Parse(content)
	if err != nil {
		log.Fatalf("Error while parsing configuration: %s", err.Error())
	}

	run(c)
}

func run(c config.Config) {
	type compiledAlias struct {
		Alias *regexp.Regexp
		Name  string
	}

	// Map every regular expression string (alias) to a compiled regular expression and the name of the locale.
	cache := make(map[string]compiledAlias)
	for name, locale := range c.Locales {
		alias, err := regexp.Compile(locale.Alias)
		if err != nil {
			log.Fatalf("Regular expression could not be compiled: %s", err.Error())
		}
		cache[locale.Alias] = compiledAlias{alias, name}
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		acceptlang := r.Header.Get("Accept-Language")
		if acceptlang == "" {
			log.Printf("No Accept-Language header has been provided\n")
			redirect(w, r, c.Locale, c.Dest)
			return
		}

		tags, q, err := language.ParseAcceptLanguage(acceptlang)
		if err != nil {
			log.Printf("Accept-Language header could not be parsed: %s\n", err.Error())
			redirect(w, r, c.Locale, c.Dest)
			return
		}

		for key, tag := range tags {
			log.Printf("%s: %f\n", tag, q[key])

			for name, locale := range c.Locales {
				var dest string
				if locale.Dest != "" {
					dest = locale.Dest
				} else {
					dest = c.Dest
				}

				if name == tag.String() || (locale.Alias != "" && cache[locale.Alias].Alias.MatchString(tag.String())) {
					redirect(w, r, name, dest)
					return
				}
			}
		}
	})

	var addr string
	if c.Network.Addr == "" {
		addr = "localhost:10451"
	} else {
		addr = c.Network.Addr
	}
	http.ListenAndServe(addr, nil)
}

func redirect(w http.ResponseWriter, r *http.Request, name string, dest string) {
	desturl := *r.URL
	path := desturl.EscapedPath()

	if path == "" || path[0] != '/' {
		path = strings.Join([]string{"/", path}, "")
	}

	mapper := func(p string) string {
		switch p {
		case "locale":
			return url.PathEscape(name)
		case "path":
			return path
		}

		return ""
	}

	desturl.RawPath = ""

	var err error
	desturl.Path, err = url.PathUnescape(os.Expand(dest, mapper))
	if err != nil {
		log.Printf("The URL path could not be unescaped: %s\n", err.Error())
	}

	log.Printf("Redirecting %s to %s\n", r.URL.String(), desturl.String())
	http.Redirect(w, r, desturl.String(), 302)
}
