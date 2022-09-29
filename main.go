package main

import (
	"log"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"strings"

	"github.com/urfave/cli/v2"
	"golang.org/x/text/language"

	"github.com/lukaswrz/relocale/config"
)

type options struct {
	Config string `short:"c" long:"config" description:"Configuration file" value-name:"FILE"`
}

func main() {
	var cf string

	app := &cli.App{
		Name:  "relocale",
		Usage: "redirect requests via the HTTP Accept-Language header",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name:        "config",
				Aliases:     []string{"c"},
				Usage:       "configuration file",
				Destination: &cf,
			},
		},
		Action: func(c *cli.Context) error {
			return nil
		},
	}

	app.Setup()

	if err := app.Run(os.Args); err != nil {
		log.Fatalf("Error: %s", err.Error())
	}

	if cf == "" {
		cf = config.Locate()

		if cf == "" {
			log.Fatal("Unable to locate configuration file")
		}
	} else {
		_, err := os.Stat(cf)
		if err != nil {
			log.Fatalf("Specified configuration file not accessible: %s", err.Error())
		}
	}

	content, err := os.ReadFile(cf)
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
