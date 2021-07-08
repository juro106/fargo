package main

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
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
	// "net/http"
	// "net/url"
	// "os/exec"

	// "bufio"
	_ "io/ioutil"
	_ "reflect"

	// "github.com/flosch/pongo2"

	// "github.com/russross/blackfriday/v2"
	"github.com/fsnotify/fsnotify"
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
	DateP         string        `yaml:"dateP" json: "dateP"`
	DatePublished string        `yaml:"datePublished" json:"datePublished"` //
	DateM         string        `yaml:"dateM" json: "dateM"`
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
	Baseurl     string   `yaml:"baseurl" json:"baseurl"`
	Title       string   `yaml:"title" json:"title"`
	Source      string   `yaml:"source" json:"source"`
	Name        string   `yaml:"name" json:"name"`
	Dst         string   `yaml:"dst" json:"dst"`
	Posts       string   `yaml:"posts" json:"posts"`
	Includes    string   `yaml:"includes" json:"includes"`
	Layouts     string   `yaml:"layouts" json:"layouts"`
	Permalink   string   `yaml:"permalink" json:"permalink"`
	Exclude     []string `yaml:"exclude" json:"exclude"`
	Host        string   `yaml:"host" json:"host"`
	Port        int      `yaml:"port" json:"port"`
	LimitPosts  int      `yaml:"limit_posts" json:"limit_posts"`
	Assets      string   `yaml:"assets"`
	MarkdownExt string   `yaml:"markdown_ext" json:"markdown_ext"`
}

type PageB2Tag map[string][]Item
type HtmlMap map[string]template.HTML

var (
	// config
	cfg Config
	now string = time.Now().Format("2006-01-02T15:04:05Z07:00")
	// 全ページのデータリスト
	metalist = make([]Meta, 0)
	// 投稿ページのリスト
	postlist = make([]Meta, 0)
	// 固定ページのリスト
	pagelist = make([]Meta, 0)
	// タグページのリスト
	tagplist = make([]Meta, 0)
	// json用
	forjsonlist = make([]PostData, 0)
	// tag リスト
	taglist []string
	// page-belong-to-tag リスト
	pageB2taglist PageB2Tag = PageB2Tag{}
	// map[string][]string = map[string][]string{}
	pB2tMap HtmlMap = HtmlMap{}
	// link リスト
	link2pMap map[string][]template.HTML = map[string][]template.HTML{}
	// assets
	css      string
	mainJs   string
	tagJs    string
	topJs    string
	isServer bool
	// server 用 html の map
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

// ページ作成の準備、タグページや関連リンク作成用の下準備
func (cfg *Config) collectData(dirName string) {
	// ディレクトリのファイル一覧を得る
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
	// dirName内のファイルをループ
	for _, file := range files {
		slots <- struct{}{}
		wg.Add(1)
		go func(file2 string) {
			fpath := file2
			// データ登録用の構造体を用意
			var meta Meta
			// サイトの基本情報を metadata に加える
			mu.Lock()
			meta.Baseurl = cfg.Baseurl
			meta.IsServer = isServer
			meta.Css = css
			meta.Now = now

			fi
			// ファイルの中身を読み取る
			srcFile := filepath.Join(dirName, fpath)

			b, e := os.ReadFile(srcFile)
			if e != nil {
				log.Fatal(e)
			}
			// 一旦 string型にして、frontmatter(metadata)を抜く
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
			// frontmatter を取り除いた部分を html に変換
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
			// markdown から html へ変換したものを Body へ登録
			body := buf.Bytes()
			meta.Body = template.HTML(body)
			// frontmatter を取得
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
			// slug を登録
			slug := filepath.Base(fpath[:len(fpath)-len(filepath.Ext(fpath))])
			if slug == "index" {
				meta.Slug = "/"
				meta.Permalink = cfg.Baseurl
			} else {
				meta.Slug = slug
				meta.Permalink = urlJoin(cfg.Baseurl, slug)
			}
			// 日付変換
			if len(meta.DatePublished) > 0 {
				meta.DateP = meta.DatePublished[:10]
			}
			if len(meta.DateModified) > 0 {
				meta.DateM = meta.DateModified[:10]
			}
			// デプロイ先の登録
			dst := "./_site"
			if fpath == "index.md" && dirName == "./_pages" {
				meta.Dst = filepath.Join(dst, "index.html")
			} else {
				dstDir := filepath.Join(dst, slug)
				if !isServer {
					err = os.MkdirAll(dstDir, 0755)
				}
				meta.Dst = filepath.Join(dst, slug, "index.html")
			}
			// tag が1つ以上あったら taglist に追加
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
			// 他のページへのリンクが1つ以上あったら link2pMap へ追加
			if len(meta.Links) > 0 {
				for _, link := range meta.Links {
					str := "<li class=\"link-list-item\"><a class=\"l-a\"href=\"/" + meta.Slug + "/\">" + meta.Title + "</a></li>"
					link2pMap[link] = append(link2pMap[link], template.HTML(str))
				}
			}
			// リストへデータ追加
			if dirName == "./_pages" {
				meta.Js = topJs
				pagelist = append(pagelist, meta)
			} else if dirName == "./_posts" {
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
			} else if dirName == "./_tags" {
				meta.Js = tagJs
				tagplist = append(tagplist, meta)
			}
			// 全ページデータへ追加
			metalist = append(metalist, meta)
			mu.Unlock()

			<-slots
			wg.Done()
		}(file)
	}
	wg.Wait()
}

// ファイルを作成して書き込み
func (cfg *Config) convertFile(tpl, ptype string) {
	var list []Meta
	if ptype == "page" {
		list = pagelist
	} else if ptype == "post" {
		list = postlist
	} else if ptype == "tag" {
		list = tagplist
	}
	// テンプレートの読み込み
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
			// テンプレートへ反映させて書き込み
			new_buf := new(bytes.Buffer)
			if err := t.Execute(new_buf, meta2); err != nil {
				log.Println("create file", err)
			}
			if !isServer {
				os.WriteFile(meta2.Dst, new_buf.Bytes(), 0644)
				fmt.Printf("%s WriteFile完了 ======>>>\n", meta2.Dst)
			}
			b := fmt.Sprintf("%v\n", new_buf)
			htmlMap[meta2.Slug] = template.HTML(b)
			// fmt.Printf("WriteData: %s\n", meta2.Slug)

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
	isServer = true
	htmlMap = map[string]template.HTML{}
}

func build(server bool) {
	isServer = server
	// fmt.Printf("cfg: %+v\n", cfg)
	// config を読み込み
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
		// css, js の hash 生成
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
	// データ収集
	cfg.collectData("./_pages")
	cfg.collectData("./_posts")
	// tag mdファイル生成
	// ① taglistの重複を削除する
	taglistM := make(map[string]struct{})
	tagList := make([]string, 0)

	for _, elem := range taglist {
		// mapの第2引数には、その値が入っているかどうかの真偽値が入っている。
		if _, ok := taglistM[elem]; !ok && len(elem) != 0 {
			taglistM[elem] = struct{}{}
			tagList = append(tagList, elem)
		}
	}
	// ② tag ベースになる .md ファイルを生成（既にあるものはスルー）※postsdataを生成してから
	dirName := "./_tags/"
	for _, tag := range tagList {
		s := []string{tag, ".md"}
		fName := strings.Join(s, "")
		srcFile := filepath.Join(dirName, fName)
		if !fileExists(srcFile) {
			copyFile("./_layouts/tag.md", srcFile)
		}
	}
	cfg.collectData("./_tags")

	// 更新日順に並べる 一覧表示用
	sort.Slice(postlist, func(i, j int) bool {
		return time2int(postlist[i].DateModified) > time2int(postlist[j].DateModified)
	})
	// 更新日順に並べる 一覧表示用
	sort.Slice(metalist, func(i, j int) bool {
		return time2int(metalist[i].DateModified) > time2int(metalist[j].DateModified)
	})
	// 更新日順に並べる 一覧表示用
	sort.Slice(forjsonlist, func(i, j int) bool {
		return time2int(forjsonlist[i].DateModified) > time2int(forjsonlist[j].DateModified)
	})

	// pages ページ生成
	cfg.convertFile("./_layouts/default.html", "page")
	// tag ページ生成
	cfg.convertFile("./_layouts/tag.html", "tag")
	// posts ページ生成
	cfg.convertFile("./_layouts/post.html", "post")

	// json化 javascript で扱うときのため
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
		copyFile("./_assets/js/prism.js", "./_site/js/prism.js")
		copyFile("./_assets/css/style.css", "./_site/"+css)
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

func checkFiles(dir string) bool {
	fs, err := os.ReadDir(dir)
	if err != nil {
		log.Fatal(err)
	}
	for _, f := range fs {
		if strings.HasSuffix(f.Name(), "~") {
			fmt.Println("file:", f.Name())
			return false
		}
	}
	return true
}

type FileInfo struct {
	Path string
	Time time.Time
}

func getpath(args []string) ([]FileInfo, error) {
	var list []FileInfo
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
	flag := false
	for i := range *list {
		if time, _ := os.Stat((*list)[i].Path); time.ModTime().After((*list)[i].Time) {
			// (*list)[i].Time = time.ModTime()
			fmt.Println((*list)[0])
			flag = true
			break
		}
	}
	return flag
}

func Watch(args []string, flg chan bool, callback func()) {
	fmt.Println("watch start!!!")
	list := []FileInfo{}
	list, err := getpath(args)
	if err != nil {
		fmt.Println("error", err)
	}
	for range time.Tick(1 * time.Second) {
		if isChanged(&list) {
			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				fmt.Println("changed")
				flg <- true
				wg.Done()
			}()
			wg.Wait()
			callback()
		}
	}
}

var watching = false

var ws2 *websocket.Conn

func rebuild(done chan bool) {
	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() {
		log.Println("clear Maps")
		clearMap()
		wg.Done()
	}()
	wg.Wait()
	time.Sleep(100 * time.Millisecond)
	wg.Add(1)
	go func() {
		log.Println("start re build")
		build(true)
		wg.Done()
	}()
	wg.Wait()
	watching = false
	reloade(ws2)
	// <-done
}
func reloade(ws *websocket.Conn) {
	err := ws.WriteMessage(1, []byte("1"))
	if err != nil {
		log.Println("WebSocket Error: ", err)
	}
}

func wsEndpoint(w http.ResponseWriter, r *http.Request) {
	log.Println("watching?", watching)
	// upgrade this connection to a WebSocket
	fmt.Println("Try Connect Websocket")
	// ws type is *websocket.Conn
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}
	ws2 = ws
	done := make(chan bool)

	if !watching {

		Notify(done, ws)
	}
	// q := make(chan bool, 1)
	// dirs := []string{"./_posts", "./_pages", "./_tags", "./_assets/css", "./_assets/js"}
	// // dirs := []string{"./_posts"}
	// // dirs := []string{"./"}
	//
	// go func() {
	// loop:
	// 	for {
	// 		select {
	// 		case <-q:
	// 			// Wait しないと上手くリロードできない
	// 			wg := &sync.WaitGroup{}
	// 			wg.Add(1)
	// 			go func() {
	// 				log.Println("clear Maps")
	// 				clearMap()
	// 				wg.Done()
	// 			}()
	// 			wg.Wait()
	// 			time.Sleep(100 * time.Millisecond)
	// 			wg.Add(1)
	// 			go func() {
	// 				log.Println("start re build")
	// 				build(true)
	// 				wg.Done()
	// 			}()
	// 			wg.Wait()
	// 			time.Sleep(100 * time.Millisecond)
	// 			err = ws.WriteMessage(1, []byte("1"))
	// 			if err != nil {
	// 				log.Println("WebSocket Error: ", err)
	// 			}
	// 			break loop
	// 			// case <-time.After(10 * time.Millisecond):
	// 			// 	break loop
	// 		}
	// 	}
	// }()
	//
	// Watch(dirs, q, func() {})
	// wg := &sync.WaitGroup{}
	// wg.Add(1)
	// go func() {
	//     reader(ws)
	//     wg.Done()
	// }()
	// wg.Wait()
}
func Notify(done chan bool, ws *websocket.Conn) {
	fmt.Println("Notify Start")
	watching = true
	counter := 0
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		log.Fatal(err)
	}
	defer watcher.Close()
	go func(done chan bool) {
		log.Println("fsnotify loop start")
		for {
			select {
			case event := <-watcher.Events:
				// log.Println("event: ", event)
				switch {
				case event.Op&fsnotify.Write == fsnotify.Write:
					log.Println("Modified file: ", event.Name)
				case event.Op&fsnotify.Create == fsnotify.Create:
					log.Println("Created file: ", event.Name)
				case event.Op&fsnotify.Rename == fsnotify.Rename:
					log.Println("Renamed file: ", event.Name)
				case event.Op&fsnotify.Chmod == fsnotify.Chmod:
					log.Println("File changed permission: ", event.Name)
				case event.Op&fsnotify.Remove == fsnotify.Remove:
					log.Println("Removed file: ", event.Name)
					if strings.Contains(event.Name, "~") {
						counter++
					}
				}
			case err := <-watcher.Errors:
				log.Println("fsnotify error: ", err)
				done <- true
			}
			if counter == 1 {
				time.Sleep(300 * time.Millisecond)
				fmt.Println("catch channel")
				watching = false
				rebuild(done)
				return
			}
		}
	}(done)
	err = watcher.Add("./_posts")
	if err != nil {
		log.Fatal(err)
	}
	<-done
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

func main() {
	cfg.load("config.yaml")
	build(true)
	fmt.Fprintf(os.Stderr, "Listening at %s:%d\n", cfg.Host, cfg.Port)
	// setupRoutes()

	r := mux.NewRouter()
	r.HandleFunc("/ws", wsEndpoint)
	r.HandleFunc("/{key}/ws", wsEndpoint)
	r.HandleFunc("/", indexHandler)
	r.HandleFunc("/{key}/", myHandler)
	r.PathPrefix("/").Handler(http.FileServer(http.Dir(cfg.Assets)))
	log.Fatal(http.ListenAndServe(fmt.Sprintf("%s:%d", cfg.Host, cfg.Port), r))
}
