package main

import (
	"fmt"
	"net/url"
	"os"
	"runtime"
	"strings"
)

type Config struct {
	MaxAgeInDays   int      `yaml:"max-age" description:"Max age of collected files; older files are ignored." shorthand:"a"`
	RootPaths      []string `yaml:"path" description:"Root paths from where files should be collected.\nSpecify multiple root paths by using this flag multiple times." shorthand:"p"`
	FileExtensions []string `yaml:"extensions" description:"File extensions that should be collected. If left empty, all files are collected.\nSpecify multiple extensions by using this flag multiple times." shorthand:"e"`
	Server         string   `yaml:"thunderstorm-server" shorthand:"s" description:"Thunderstorm URL to which files should be uploaded.\nExample: --thunderstorm-server https://my.thunderstorm:8080/"`
	Sync           bool     `yaml:"upload-synchronous" description:"Whether files should be uploaded synchronously to Thunderstorm. If yes, the collector takes longer, but displays the results of all scanned files."`
	Debug          bool     `yaml:"debug" description:"Print debugging information." hidden:"true"`
	Threads        int      `yaml:"threads" description:"How many threads should upload files simultaneously." shorthand:"r"`
	MaxFileSize    int64    `yaml:"max-filesize" description:"Maximum file size up to which files should be uploaded (in MB)." shorthand:"m"`
	Proxy          string   `yaml:"http-proxy" description:"Proxy that should be used for the connection to Thunderstorm.\nIf left empty, the proxy is filled from the HTTP_PROXY and HTTPS_PROXY environment variables."`
	CAs            []string `yaml:"ca" description:"Path to a PEM CA certificate that signed the HTTPS certificate of the Thunderstorm server.\nSpecify multiple CAs by using this flag multiple times."`
	Insecure       bool     `yaml:"insecure" description:"Don't verify the Thunderstorm certificate if HTTPS is used."`
	Logfile        string   `yaml:"logfile" description:"Write the log to this file as well as to the console." shorthand:"l"`
	Source         string   `yaml:"source" description:"Name for this device in the Thunderstorm log messages." shorthand:"o"`
	Template       string   `flag:"template" description:"Process default scan parameters from this YAML file." shorthand:"t"`
	Help           bool     `flag:"help" description:"Show this help." shorthand:"h"`
}

var DefaultConfig = Config{
	Threads:     1,
	MaxFileSize: 100,
	RootPaths:   []string{getRootPath()},
	Source:      HostnameOrBlank(),
}

func HostnameOrBlank() string {
	hostname, _ := os.Hostname()
	return hostname
}

func getRootPath() string {
	if runtime.GOOS == "windows" {
		return "C:\\"
	} else {
		return "/"
	}
}

func ParseConfig() Config {
	var config = DefaultConfig
	err := ReadTemplateFile(&config)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	flags := CreateFlagset(&config)
	flags.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTION]...\n", os.Args[0])
		flags.PrintDefaults()
	}
	err = flags.Parse(os.Args)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		flags.Usage()
		os.Exit(1)
	}
	if config.Help || len(os.Args) == 1 {
		flags.Usage()
		os.Exit(0)
	}
	if config.Server == "" {
		fmt.Fprintln(os.Stderr, "Thunderstorm Server URL not specified")
		os.Exit(1)
	}
	if config.Threads < 1 {
		fmt.Fprintln(os.Stderr, "Thread count must be > 0")
		os.Exit(1)
	}
	if config.MaxAgeInDays < 0 {
		fmt.Fprintln(os.Stderr, "Maximum file age must be >= 0")
		os.Exit(1)
	}
	if config.MaxFileSize < 1 {
		fmt.Fprintln(os.Stderr, "Maximum file size must be >= 0")
		flags.Usage()
		os.Exit(1)
	}
	config.Server = strings.TrimSuffix(config.Server, "/")

	if !(strings.HasPrefix(config.Server, "http://") || strings.HasPrefix(config.Server, "https://")) {
		fmt.Fprintln(os.Stderr, "Missing http:// or https:// prefix in Thunderstorm URL")
		os.Exit(1)
	}

	if _, err := url.Parse(config.Server); err != nil {
		fmt.Fprintln(os.Stderr, "URL for Thunderstorm is invalid")
		os.Exit(1)
	}
	return config
}
