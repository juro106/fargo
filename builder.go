package main

import (
	"fmt"
	"log"
    _ "bytes"
    // "bufio"
	// "math"
	"net/http"
	// "net/url"
    // "reflect"
	_ "io"
	"os"
    _ "regexp"
    _ "io/ioutil"
	// "os/exec"
	_ "path/filepath"
	_ "path"
	// "runtime"
	_ "sort"
	_ "strconv"
	_ "strings"
	_ "time"
    _ "html/template"
    _ "encoding/json"
    _ "sync"
    _ "crypto/md5"
    _ "encoding/hex"

	// "github.com/flosch/pongo2"
	// "github.com/howeyc/fsnotify"
	// "github.com/russross/blackfriday/v2"
    _ "github.com/yuin/goldmark"
    _ "github.com/yuin/goldmark/extension"
    _ "github.com/yuin/goldmark/renderer/html"
    // "github.com/yuin/goldmark/parser"
    // "github.com/yuin/goldmark-highlighting"
	"gopkg.in/yaml.v2")

type Config struct {
    Baseurl     string                       `yaml:"baseurl" json:"baseurl"`
    Title       string                       `yaml:"title" json:"title"`
    Source      string                       `yaml:"source" json:"source"`
    Name        string                       `yaml:"name" json:"name"`
    Dst         string                       `yaml:"dst" json:"dst"`
    Posts       string                       `yaml:"posts" json:"posts"`
    Data        string                       `yaml:"data" json:"data"`
    Includes    string                       `yaml:"includes" json:"includes"`
    Layouts     string                       `yaml:"layouts" json:"layouts"`
    Permalink   string                       `yaml:"permalink" json:"permalink"`
    Exclude     []string                     `yaml:"exclude" json:"exclude"`
    Host        string                       `yaml:"host" json:"host"`
    Port        int                          `yaml:"port" json:"port"`
    LimitPosts  int                          `yaml:"limit_posts" json:"limit_posts"`
    MarkdownExt string                       `yaml:"markdown_ext" json:"markdown_ext"`
}

func checkFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func (cfg *Config) load(file string) error {
    b, err := os.ReadFile(file)
    if err != nil {
        return err
    }

    err = yaml.Unmarshal(b, cfg)
    if err != nil {
        return err
    }
    return nil
}

func (cfg *Config) Serve() error {

    // checkFatal(err)
    fmt.Fprintf(os.Stderr, "Lisning at %s:%d\n", cfg.Host, cfg.Port)
    return http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),  http.FileServer(http.Dir(cfg.Dst)))
}

