package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	// "math"
	// "net/url"
	// "os/exec"
	// "runtime"
	// "bufio"
	_ "io/ioutil"
	_ "reflect"

	"fargo/ffsnotify"
	// "github.com/russross/blackfriday/v2"
	// "github.com/dietsche/rfsnotify"
	"github.com/yuin/goldmark"
	_ "github.com/yuin/goldmark"
	"github.com/yuin/goldmark/extension"
	_ "github.com/yuin/goldmark/extension"
	"github.com/yuin/goldmark/renderer/html"
	_ "github.com/yuin/goldmark/renderer/html"

	// "github.com/yuin/goldmark/parser"
	// "github.com/yuin/goldmark-highlighting"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"

	"github.com/fsnotify/fsnotify"
	"gopkg.in/yaml.v2"
)

type Item struct {
	Title       string        `json:"title"`
	Slug        string        `json:"slug"`
	SummaryHTML template.HTML `json:"summary"`
	Thumnail    bool          `json:"thumnail"`
}

type Meta struct {
	Title         string        `yaml:"title" json:"title"`             // meta
	Description   string        `yaml:"description" json:"description"` //
	Tags          []string      `yaml:"tags" json:"tags"`               //
	Links         []string      `yaml:"links" json:"links"`             //
	DatePub       string        `yaml:"datePub" json: "datePub"`
	DatePublished string        `yaml:"datePublished" json:"datePublished"` //
	DateMod       string        `yaml:"dateMod" json: "dateMod"`
	DateModified  string        `yaml:"dateModified" json:"dateModified"` //
	Draft         bool          `yaml:"draft" json:"draft"`               //
	Home          bool          `yaml:"home" json:"home"`                 //
	Fixed         bool          `yaml:"fixed" json:"fixed"`               //
	Code          bool          `yaml:"code" json:"code"`                 //
	Option        []string      `yaml:"option" json:"option"`             //
	Layout        string        `yaml:"layout" json:"layout"`             //
	Slug          string        `yaml:"slug" json:"slug"`                 //
	Permalink     string        `yaml:"permalink" json:"permalink"`       //
	Summary       string        `yaml:"summary" json:"summary"`           //
	SummaryHTML   template.HTML `yaml:"summaryHTML" json:"summaryHTML"`   //
	Thumnail      bool
	Dst           string
	Body          template.HTML
	PageTag       PageB2Tag
	B2Page        []Item
	B2Pagelen     int
	Link2P        []template.HTML
	Plist         []Meta
	Plistlen      int
	Baseurl       string
	PB2T          []template.HTML
	Now           string
	Css           string
	Js            string
	IsServer      bool
}

type PostData struct {
	Title         string        `json:"title"`         // meta
	Tags          []string      `json:"tags"`          //
	DatePublished string        `json:"datePublished"` //
	DateModified  string        `json:"dateModified"`  //
	Slug          string        `json:"slug"`          //
	Summary       template.HTML `json:"summary"`       //
	Thumnail      bool          `json:"thumnail"`
}

type Config struct {
	Baseurl     string   `yaml:"baseurl"`
	Title       string   `yaml:"title"`
	Source      string   `yaml:"source"`
	Name        string   `yaml:"name"`
	Dst         string   `yaml:"dst"`
	Pages       string   `yaml:"pages"`
	Posts       string   `yaml:"posts"`
	Tags        string   `yaml:"tags"`
	Includes    string   `yaml:"includes"`
	Layouts     string   `yaml:"layouts"`
	Permalink   string   `yaml:"permalink"`
	Exclude     []string `yaml:"exclude"`
	Watch       []string `yaml:"watch"`
	Host        string   `yaml:"host"`
	Port        int      `yaml:"port"`
	LimitPosts  int      `yaml:"limit_posts"`
	Assets      string   `yaml:"assets"`
	MarkdownExt string   `yaml:"markdown_ext"`
}

type PageB2Tag map[string][]Item
type HtmlMap map[string]template.HTML

var (
	// config
	cfg Config
	now string = time.Now().Format("2006-01-02T15:04:05Z07:00")
	// ?????????????????????????????????
	metalist = make([]Meta, 0)
	// ???????????????????????????
	postlist = make([]Meta, 0)
	// ???????????????????????????
	pagelist = make([]Meta, 0)
	// ???????????????????????????
	tagplist = make([]Meta, 0)
	// json???
	forjsonlist = make([]PostData, 0)
	// tag ?????????
	taglist []string
	// page-belong-to-tag ?????????
	pageB2taglist PageB2Tag = PageB2Tag{}
	// map[string][]string = map[string][]string{}
	pB2tMap HtmlMap = HtmlMap{}
	// link ?????????
	link2pMap map[string][]template.HTML = map[string][]template.HTML{}
	// assets
	css      string
	mainJs   string
	tagJs    string
	topJs    string
	isServer bool
	// server ??? html ??? map
	htmlMap map[string]template.HTML = map[string]template.HTML{}
)

func checkFatal(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func copyFile(src, dst string) (int64, error) {
	sf, err := os.Open(src)
	if err != nil {
		return 0, err
	}
	defer sf.Close()
	df, err := os.Create(dst)
	if err != nil {
		return 0, err
	}
	defer df.Close()
	return io.Copy(df, sf)
}

func imgCopy(src, dst string) {
	fs, err := os.ReadDir(src)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range fs {
		srcPath := filepath.Join(src, f.Name())
		dstPath := filepath.Join(dst, f.Name())
		if f.IsDir() {
			imgCopy(srcPath, dst)
			continue
		}
		fInfo, err := f.Info()
		if err != nil {
			log.Fatal(err)
		}
		df, _ := os.Open(dstPath)
		defer df.Close()
		if fi, err := df.Stat(); err == nil {
			if fi.Size() == fInfo.Size() {
				continue
			}
		}
		copyFile(srcPath, dstPath)
		fmt.Printf("cp %s >>> %s\n", srcPath, dstPath)
	}
}

func time2int(args interface{}) int {
	dateTime := args.(string)
	var i int
	dateTime = strings.Replace(dateTime, "-", "", -1)
	dateTime = strings.Replace(dateTime, ":", "", -1)
	dateTime = strings.Replace(dateTime, "T", "", -1)
	dateTime = strings.Replace(dateTime, "+", "", -1)
	i, _ = strconv.Atoi(dateTime)
	return i
}

func str(s interface{}) string {
	if ss, ok := s.(string); ok {
		return ss
	}
	return ""
}

func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}

func urlJoin(l, r string) string {
	r = path.Clean(r)
	ls := strings.HasSuffix(l, "/")
	rp := strings.HasPrefix(r, "/")

	if ls && rp {
		return l + r[1:] + "/"
	}
	if !ls && !rp {
		return l + "/" + r + "/"
	}
	return l + r + "/"
}

func removeTag(str string) string {
	rep := regexp.MustCompile(`<("[^"]*"|'[^']*'|[^'">])*>`)
	str = rep.ReplaceAllString(str, "")
	return str
}

func sliceStr(str string, num int) string {
	if len(str) > num {
		return str[:num]
	}
	return str
}

func hash_file_md5(filePath string) (string, error) {
	var returnMD5String string
	file, err := os.Open(filePath)
	if err != nil {
		return returnMD5String, err
	}
	defer file.Close()
	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return returnMD5String, err
	}
	hashInBytes := hash.Sum(nil)[:16]
	returnMD5String = hex.EncodeToString(hashInBytes)
	return returnMD5String, nil
}

func hash_name(filePath, fileName, ext string) (string, error) {
	hash, err := hash_file_md5(filePath)
	if err != nil {
		fmt.Println(hash, err)
	}
	return fileName + hash + ext, nil
}

func hash_params(filePath, fileName string) string {
	hash, err := hash_file_md5(filePath)
	if err != nil {
		fmt.Println(hash, err)
	}
	return fileName + "?" + hash
}

// ?????????????????????????????????????????????????????????????????????????????????
func (cfg *Config) collectData(dirName string) {
	// ????????????????????????????????????????????????
	// files, err := os.ReadDir(dirName)
	f, err := os.Open(dirName)
	if err != nil {
		log.Fatal(err)
	}
	defer f.Close()
	files, _ := f.Readdirnames(0)

	limit := 1
	slots := make(chan struct{}, limit)
	var mu sync.RWMutex
	var wg sync.WaitGroup
	// dirName??????????????????????????????
	for _, file := range files {
		slots <- struct{}{}
		wg.Add(1)
		go func(file2 string) {
			fpath := file2
			// ???????????????????????????????????????
			var meta Meta
			// ??????????????????????????? metadata ????????????
			mu.Lock()
			meta.Baseurl = cfg.Baseurl
			meta.IsServer = isServer
			meta.Css = css
			meta.Now = now

			// ????????????????????????????????????
			srcFile := filepath.Join(dirName, fpath)

			b, e := os.ReadFile(srcFile)
			if e != nil {
				log.Fatal(e)
			}
			// ?????? string???????????????frontmatter(metadata)?????????
			content := string(b)
			lines := strings.Split(content, "\n")
			if len(lines) > 2 && lines[0] == "---" {
				var n int
				var line string
				for n, line = range lines[1:] {
					if line == "---" {
						break
					}
				}
				content = strings.Join(lines[n+2:], "\n")
			}
			// frontmatter ??????????????????????????? html ?????????
			md := goldmark.New(
				goldmark.WithExtensions(extension.GFM),
				goldmark.WithRendererOptions(
					html.WithUnsafe(),
				),
			)
			var buf bytes.Buffer
			if err = md.Convert([]byte(content), &buf); err != nil {
				panic(err)
			}
			// markdown ?????? html ???????????????????????? Body ?????????
			body := buf.Bytes()
			meta.Body = template.HTML(body)
			// frontmatter ?????????
			err = yaml.Unmarshal([]byte(b), &meta)
			if err != nil {
				log.Fatalf("error: %v", err)
			}
			if len(meta.Summary) == 0 {
				meta.SummaryHTML = template.HTML(meta.Description)
			} else {
				meta.SummaryHTML = template.HTML(strings.Replace(meta.Summary, "\n", "<br>", -1))
				if strings.Contains(meta.Summary, "<img") {
					meta.Thumnail = true
				}
			}
			// slug ?????????
			slug := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
			if slug == "index" {
				meta.Slug = "/"
				meta.Permalink = cfg.Baseurl
			} else {
				meta.Slug = slug
				meta.Permalink = urlJoin(cfg.Baseurl, slug)
			}
			// ????????????
			if len(meta.DatePublished) > 0 {
				meta.DatePub = meta.DatePublished[:10]
			}
			if len(meta.DateModified) > 0 {
				meta.DateMod = meta.DateModified[:10]
			}
			// ????????????????????????
			dst := cfg.Dst
			if fpath == "index.md" && dirName == cfg.Pages {
				meta.Dst = filepath.Join(dst, "index.html")
			} else {
				dstDir := filepath.Join(dst, slug)
				if !isServer {
					err = os.MkdirAll(dstDir, 0755)
				}
				meta.Dst = filepath.Join(dst, slug, "index.html")
			}
			// tag ???1????????????????????? taglist ?????????
			if len(meta.Tags) > 0 {
				// page-tag-list
				taglist = append(taglist, meta.Tags...)
				// tag-page-list
				for _, tag := range meta.Tags {
					var item Item
					item.Title = meta.Title
					item.Slug = meta.Slug
					item.SummaryHTML = meta.SummaryHTML
					item.Thumnail = meta.Thumnail
					pageB2taglist[tag] = append(pageB2taglist[tag], item)
				}
			}
			// ?????????????????????????????????1????????????????????? link2pMap ?????????
			if len(meta.Links) > 0 {
				for _, link := range meta.Links {
					str := "<li class=\"link-list-item\"><a class=\"l-a\"href=\"/" + meta.Slug + "/\">" + meta.Title + "</a></li>"
					link2pMap[link] = append(link2pMap[link], template.HTML(str))
				}
			}
			// ???????????????????????????
			if dirName == cfg.Pages {
				meta.Js = topJs
				pagelist = append(pagelist, meta)
			} else if dirName == cfg.Posts {
				meta.Js = mainJs
				postlist = append(postlist, meta)
				var pmeta PostData
				pmeta.Title = meta.Title
				pmeta.Tags = meta.Tags
				pmeta.DatePublished = meta.DatePublished
				pmeta.DateModified = meta.DateModified
				pmeta.Slug = meta.Slug
				pmeta.Summary = meta.SummaryHTML
				pmeta.Thumnail = meta.Thumnail
				forjsonlist = append(forjsonlist, pmeta)
			} else if dirName == cfg.Tags {
				meta.Js = tagJs
				tagplist = append(tagplist, meta)
			}
			// ??????????????????????????????
			metalist = append(metalist, meta)
			mu.Unlock()

			<-slots
			wg.Done()
		}(file)
	}
	wg.Wait()
}

// ???????????????????????????????????????
func (cfg *Config) convertFile(tpl, ptype string) {
	var list []Meta
	if ptype == "page" {
		list = pagelist
	} else if ptype == "post" {
		list = postlist
	} else if ptype == "tag" {
		list = tagplist
	}
	// ?????????????????????????????????
	t := template.Must(template.ParseFiles(tpl, "./_layouts/footer.html"))

	var wg sync.WaitGroup
	var mu sync.RWMutex

	for _, meta := range list {
		wg.Add(1)
		go func(meta2 Meta) {
			mu.Lock()
			if ptype == "page" && meta2.Slug == "/" {
				meta2.Plist = postlist
				meta2.Plistlen = len(meta2.Plist)
			}
			if ptype == "tag" {
				meta2.B2Page = pageB2taglist[meta2.Slug]
				meta2.B2Pagelen = len(meta2.B2Page)
				tt := template.Must(template.ParseFiles("./_layouts/pB2t.html"))
				buf := new(bytes.Buffer)
				if err := tt.Execute(buf, meta2); err != nil {
					log.Println("create file", err)
				}
				// fmt.Printf("type of &buf: %s\n", reflect.TypeOf(buf))
				b := fmt.Sprintf("%v\n", buf)
				html := template.HTML(b)
				pB2tMap[meta2.Slug] = html
			}
			if ptype == "post" {
				for _, tag := range meta2.Tags {
					meta2.PB2T = append(meta2.PB2T, pB2tMap[tag])
				}
			}
			// links
			if len(link2pMap[meta2.Slug]) > 0 {
				meta2.Link2P = link2pMap[meta2.Slug]
			}
			// ????????????????????????????????????????????????
			new_buf := new(bytes.Buffer)
			if err := t.Execute(new_buf, meta2); err != nil {
				log.Println("create file", err)
			}
			// ?????????????????????????????????????????????????????????????????????????????????????????????
			if isServer {
				b := fmt.Sprintf("%v\n", new_buf)
				htmlMap[meta2.Slug] = template.HTML(b)
				// fmt.Printf("WriteData: %s\n", meta2.Slug)
			} else {
				os.WriteFile(meta2.Dst, new_buf.Bytes(), 0644)
				fmt.Printf("%s WriteFile?????? ======>>>\n", meta2.Dst)
			}
			mu.Unlock()
			wg.Done()
		}(meta)
	}
	wg.Wait()
}

func clearMap() {
	now = time.Now().Format("2006-01-02T15:04:05Z07:00")
	metalist = []Meta{}
	postlist = []Meta{}
	pagelist = []Meta{}
	tagplist = []Meta{}
	forjsonlist = []PostData{}
	taglist = []string{}
	pageB2taglist = PageB2Tag{}
	pB2tMap = HtmlMap{}
	link2pMap = map[string][]template.HTML{}
	htmlMap = map[string]template.HTML{}
}

func build() {
	// fmt.Printf("cfg: %+v\n", cfg)
	// config ???????????????
	buf, err := os.ReadFile("config.yaml")
	if err != nil {
		log.Fatal(err)
	}
	err = yaml.Unmarshal(buf, &cfg)
	if err != nil {
		log.Fatal(err)
	}

	if isServer {
		css = hash_params("_assets/css/style.css", "/css/style.css")
		mainJs = hash_params("_assets/js/main.js", "/js/main.js")
		tagJs = hash_params("_assets/js/tag.js", "/js/tag.js")
		topJs = hash_params("_assets/js/top-page.js", "/js/top-page.js")
	} else {
		// css, js ??? hash ??????
		css, err = hash_name("_assets/css/style.css", "/css/style-", ".css")
		if err != nil {
			log.Fatal(err)
		}
		mainJs, err = hash_name("_assets/js/main.js", "/js/main-", ".js")
		if err != nil {
			log.Fatal(err)
		}
		tagJs, err = hash_name("_assets/js/tag.js", "/js/tag-", ".js")
		if err != nil {
			log.Fatal(err)
		}
		topJs, err = hash_name("_assets/js/top-page.js", "/js/top-", ".js")
		if err != nil {
			log.Fatal(err)
		}
	}
	// ???????????????
	cfg.collectData(cfg.Pages)
	cfg.collectData(cfg.Posts)
	// tag md??????????????????
	// ??? taglist????????????????????????
	taglistM := make(map[string]struct{})
	tagList := make([]string, 0)

	for _, elem := range taglist {
		// map??????2???????????????????????????????????????????????????????????????????????????????????????
		if _, ok := taglistM[elem]; !ok && len(elem) != 0 {
			taglistM[elem] = struct{}{}
			tagList = append(tagList, elem)
		}
	}
	// ??? tag ?????????????????? .md ????????????????????????????????????????????????????????????postsdata?????????????????????
	dirName := cfg.Tags
	for _, tag := range tagList {
		s := []string{tag, ".md"}
		fName := strings.Join(s, "")
		srcFile := filepath.Join(dirName, fName)
		if !fileExists(srcFile) {
			copyFile("./_layouts/tag.md", srcFile)
		}
	}
	cfg.collectData(cfg.Tags)

	// ???????????????????????? ???????????????
	sort.Slice(postlist, func(i, j int) bool {
		return time2int(postlist[i].DateModified) > time2int(postlist[j].DateModified)
	})
	// ???????????????????????? ???????????????
	sort.Slice(metalist, func(i, j int) bool {
		return time2int(metalist[i].DateModified) > time2int(metalist[j].DateModified)
	})
	// ???????????????????????? ???????????????
	sort.Slice(forjsonlist, func(i, j int) bool {
		return time2int(forjsonlist[i].DateModified) > time2int(forjsonlist[j].DateModified)
	})

	// pages ???????????????
	cfg.convertFile("./_layouts/default.html", "page")
	// tag ???????????????
	cfg.convertFile("./_layouts/tag.html", "tag")
	// posts ???????????????
	cfg.convertFile("./_layouts/post.html", "post")

	// json??? javascript ????????????????????????
	data_json, _ := json.Marshal(&metalist)
	post_json, _ := json.Marshal(&forjsonlist)
	tags_json, _ := json.Marshal(&pageB2taglist)
	os.WriteFile("./_assets/data/page-data.json", data_json, 0644)
	os.WriteFile("./_assets/data/post-data.json", post_json, 0644)
	os.WriteFile("./_assets/data/tag-data.json", tags_json, 0644)

	if !isServer {
		dataDir := filepath.Join(cfg.Dst, "data")
		jsDir := filepath.Join(cfg.Dst, "js")
		cssDir := filepath.Join(cfg.Dst, "css")
		imgDir := filepath.Join(cfg.Dst, "img")
		err = os.MkdirAll(dataDir, 0755)
		err = os.MkdirAll(jsDir, 0755)
		err = os.MkdirAll(cssDir, 0755)
		err = os.MkdirAll(imgDir, 0755)

		cssFiles, _ := filepath.Glob("./_site/css/*")
		jsFiles, _ := filepath.Glob("./_site/js/*")
		for _, f := range cssFiles {
			if err := os.Remove(f); err != nil {
				fmt.Println(err)
			}
		}
		for _, f := range jsFiles {
			if err := os.Remove(f); err != nil {
				fmt.Println(err)
			}
		}
		copyFile("./_assets/js/main.js", "./_site/"+mainJs)
		copyFile("./_assets/js/top-page.js", "./_site/"+topJs)
		copyFile("./_assets/js/tag.js", "./_site/"+tagJs)
		copyFile("./_assets/css/style.css", "./_site/"+css)

		copyFile("./_assets/js/prism.js", "./_site/js/prism.js")
		copyFile("./_assets/ads.txt", "./_site/ads.txt")
		copyFile("./_assets/favicon.ico", "./_site/favicon.ico")
		// deploy images
		imgCopy("./_assets/img", "./_site/img")

		// fmt.Printf("[page-data.json]: %v\n", string(post_json))
		os.WriteFile("./_site/data/page-data.json", data_json, 0644)
		os.WriteFile("./_site/data/post-data.json", post_json, 0644)
		os.WriteFile("./_site/data/tag-data.json", tags_json, 0644)

		// create sitemap.xml
		sitemap := make([]string, 0)
		plist := make([]string, 0)
		for _, post := range metalist {
			p := "<url><loc>" + post.Permalink + "</loc></url>"
			plist = append(plist, p)
		}
		// fmt.Printf("plist: %s\n", plist)
		sitemap = append(sitemap, "<urlset xmlns='http://www.sitemaps.org/schemas/sitemap/0.9'>")
		// strings.Join(plist, "")
		sitemap = append(sitemap, plist...)
		sitemap = append(sitemap, "</urlset>")
		sitemapStr := strings.Join(sitemap, "\n")
		// fmt.Printf("sitemap: %s\n", sitemapStr)
		os.WriteFile("./_site/sitemap.xml", []byte(sitemapStr), 0644)
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

// ?????????????????????
var watching = false
var ws2 *websocket.Conn

// We'll need to define an Upgrader
// this will require a Read and Write buffer size
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

// define a reader which will listen for
// new messages being sent to our WebSocket
// endpoint
func reader(conn *websocket.Conn) {
	for {
		// read in a message
		messageType, p, err := conn.ReadMessage()
		if err != nil {
			log.Println(err)
			return
		}
		// print out that message for clarity
		log.Println(string(p))

		if err := conn.WriteMessage(messageType, p); err != nil {
			log.Println(err)
			return
		}

	}
}

type FileInfo struct {
	Path string
	Time time.Time
}

func getPath(args []string) ([]FileInfo, error) {
	list := []FileInfo{}
	for _, arg := range args {
		filepath.Walk(arg, func(path string, info os.FileInfo, err error) error {
			if !info.IsDir() {
				list = append(list, FileInfo{path, info.ModTime()})
			}
			return nil
		})
	}
	return list, nil
}

func isChanged(list *[]FileInfo) bool {
	flg := false
	for i := range *list {
		if time, _ := os.Stat((*list)[i].Path); time.ModTime().After((*list)[i].Time) {
			// (*list)[i].Time = time.ModTime()
			// fmt.Println((*list)[0])
			flg = true
			break
		}
	}
	return flg
}

func Watch(args []string, flg chan bool) {
	watching = true
	fmt.Println("watch start!!!")
	list, err := getPath(args)
	if err != nil {
		fmt.Println("error", err)
	}
	for range time.Tick(1 * time.Second) {
		if isChanged(&list) {
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				rebuild()
				wg.Done()
			}()
			wg.Wait()
			reload(ws2)
			return
		}
	}
}

func reload(ws *websocket.Conn) {
	watching = false
	time.Sleep(100 * time.Millisecond)
	err := ws.WriteMessage(1, []byte("1"))
	if err != nil {
		log.Println("WebSocket Error: ", err)
	}
}
func rebuild() {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		log.Println("clear Maps")
		clearMap()
		wg.Done()
	}()
	wg.Wait()
	wg.Add(1)
	go func() {
		log.Println("start re build")
		build()
		wg.Done()
	}()
	wg.Wait()

}

func fileWatcher() {
	watching = true
	watcher, err := ffsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	counter := 0

	defer watcher.Close()
	done := make(chan bool)
	go func() {
		for {
			select {
			case e := <-watcher.Events:
				switch {
				case e.Op&fsnotify.Write == fsnotify.Write:
					log.Printf("Write:  %s: %s", e.Op, e.Name)
				case e.Op&fsnotify.Create == fsnotify.Create:
					log.Printf("Create: %s: %s", e.Op, e.Name)
				case e.Op&fsnotify.Rename == fsnotify.Rename:
					log.Printf("Rename: %s: %s", e.Op, e.Name)
				case e.Op&fsnotify.Chmod == fsnotify.Chmod:
					log.Printf("Chmod:  %s: %s", e.Op, e.Name)
				case e.Op&fsnotify.Remove == fsnotify.Remove:
					log.Printf("Remove: %s: %s", e.Op, e.Name)
					counter++
				}
			case err := <-watcher.Errors:
				log.Println("rfsnotify err: ", err)
				<-done
			}
			if counter > 1 {
				wg := &sync.WaitGroup{}
				wg.Add(1)
				go func() {
					rebuild()
					wg.Done()
				}()
				wg.Wait()
				reload(ws2)
				return
			}
		}
	}()
	for _, dir := range cfg.Watch {
		watcher.AddRecursive(dir)
	}
	<-done
}

func (cfg *Config) wsEndpoint(w http.ResponseWriter, r *http.Request) {
	if watching {
		log.Println("watching->", watching, "...keep watching")
	} else {

		log.Println("watching->", watching, "...restart watching")
	}
	// upgrade this connection to a WebSocket
	fmt.Println("Try Connect Websocket")
	// ws type is *websocket.Conn
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	ws2 = ws

	log.Println("Client Connected")
	err = ws.WriteMessage(1, []byte("Hi Client!"))
	if err != nil {
		log.Println(err)
	}

	// q := make(chan bool, 1)
	// ?????????????????????????????????data ??? build ???????????????????????????????????????????????????
	// dirs := cfg.Watch
	// dirs := []string{"./_posts"}
	// dirs := []string{"./"}
	if !watching {
		fileWatcher()
		// Watch(dirs, q)
	}
}

func myHandler(w http.ResponseWriter, r *http.Request) {
	// fmt.Printf("%#v\n", r.RequestURI)
	vars := mux.Vars(r)
	// fmt.Printf("%#v\n", vars)

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, template.HTML(htmlMap[vars["key"]]))
}
func indexHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s\n", "index.html")

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, template.HTML(htmlMap["/"]))
}

func server() {
	isServer = true
	build()
	fmt.Fprintf(os.Stderr, "Listening at %s:%d\n", cfg.Host, cfg.Port)
	// setupRoutes()

	r := mux.NewRouter()
	r.HandleFunc("/ws", cfg.wsEndpoint)
	// r.HandleFunc("/{key}/ws", wsEndpoint)
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/{key}/", myHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(cfg.Assets)))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), r))
}
func main() {
	flag.Parse()
	arg := flag.Arg(0)
	switch arg {
	case "s", "serve", "server":
		server()
	case "b", "build":
		build()
	default:
		fmt.Println("please input args. 's': server, 'b': build")
	}
}
