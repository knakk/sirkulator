package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"

	"github.com/knakk/sirkulator/http"
	"golang.org/x/text/language"
)

// Main represents the execution of the sirkulatord program
type Main struct {
	Config     Config
	HTTPServer *http.Server
}

func (m Main) Run(ctx context.Context) error {
	fmt.Printf("running with config: %+v\n", m.Config)

	m.HTTPServer.Addr = fmt.Sprintf("localhost:%d", m.Config.Port)
	m.HTTPServer.Lang = m.Config.Lang
	if err := m.HTTPServer.Open(); err != nil {
		return err
	}
	return nil
}

func (m Main) Close() error {
	return nil
}

type Config struct {
	Port      int
	Lang      language.Tag
	AssetsDir string
}

func parseFlags(args []string) Config {
	conf := Config{
		Lang: language.Norwegian, // default language
	}
	fs := flag.NewFlagSet("sirkulatord", flag.ExitOnError)
	fs.Func("lang", "language: 'no' for Norwegian or 'en' for English (default: 'no')", func(s string) error {
		switch strings.ToLower(s) {
		case "no", "no-nb":
			conf.Lang = language.Norwegian
		case "en":
			conf.Lang = language.English
		default:
			return errors.New("unsupported language")
		}
		return nil
	})
	fs.IntVar(&conf.Port, "port", 9999, "port")
	fs.StringVar(&conf.AssetsDir, "assets", "", "assets directory, overriding default embedded static assets")
	fs.Parse(args)
	return conf
}

func main() {
	// Parse flags into a valid Config, will exit(1) on errors.
	conf := parseFlags(os.Args[1:])

	// Set up base context and shutdown signal handler.
	ctx, cancel := context.WithCancel(context.Background())
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt)
	go func() { <-shutdown; cancel() }()

	m := Main{
		Config:     conf,
		HTTPServer: http.NewServer(ctx, conf.AssetsDir),
	}

	// Run the program: starts all services and listeners.
	if err := m.Run(ctx); err != nil {
		m.Close()
		log.Println(err)
		os.Exit(1)
	}

	// Wait for shutdown signal.
	<-ctx.Done()

	// Clean up before exit.
	if err := m.Close(); err != nil {
		log.Println(err)
		os.Exit(1)
	}
}
