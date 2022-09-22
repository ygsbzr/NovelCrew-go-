package main

import (
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"syscall"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/lxn/win"
)

var remap map[string]string = make(map[string]string)
var docmap map[string]string = make(map[string]string)

type Book struct {
	baseburl   string
	booknumber string
}

var mainWindow *walk.MainWindow
var inTE, outTE *walk.TextEdit
var textPL *walk.TextLabel
var siteText *walk.TextLabel

func GetSystemMetrics(nIndex int) int {
	ret, _, _ := syscall.NewLazyDLL(`User32.dll`).NewProc(`GetSystemMetrics`).Call(uintptr(nIndex))
	return int(ret)
}
func main() {
	remap["www.zhhbq.com"] = `<script>[\s\S]*?</div>([\s\S]*?)<script>`
	remap["www.xbiquge.la"] = `([\s\S]*?)<p>`
	remap["www.ddyueshu.com"] = `([\s\S]*?)<p>`
	remap["www.qu-la.com"] = `([\s\S]*)`
	remap["www.88gp.net"] = `([\s\S]*)`
	remap["www.biqugesk.org"] = `([\s\S]*)`
	docmap["www.zhhbq.com"] = "div#content"
	docmap["www.xbiquge.la"] = "div#content"
	docmap["www.ddyueshu.com"] = "div#content"
	docmap["www.qu-la.com"] = "div[id=chapter-title]~div"
	docmap["www.88gp.net"] = "p#articlecontent"
	docmap["www.biqugesk.org"] = "div#booktext"
	screenX, screenY := GetSystemMetrics(0), GetSystemMetrics(1)
	width, height := 1000, 900
	sites := "支持的网址："
	for site, _ := range docmap {
		sites += site + ","
	}
	err := MainWindow{
		Title:    "简易爬虫",
		AssignTo: &mainWindow,
		Layout:   VBox{},
		Bounds:   Rectangle{Width: width, Height: height, X: (screenX - width) / 2, Y: (screenY - height) / 2},
		Children: []Widget{
			TextLabel{AssignTo: &textPL, Text: "爬小说", TextAlignment: AlignHCenterVCenter},
			HSplitter{
				Children: []Widget{
					TextEdit{AssignTo: &inTE},
					TextEdit{AssignTo: &outTE, ReadOnly: true},
				},
			},
			TextLabel{AssignTo: &siteText, Text: sites},
			PushButton{
				Text: "开始爬取",
				OnClicked: func() {
					Start()
				},
			},
		},
	}.Create()
	if err != nil {
		walk.MsgBox(nil, "错误", err.Error(), walk.MsgBoxIconError)
		return
	}
	hwhd := mainWindow.Handle()
	currstyle := win.GetWindowLong(hwhd, win.GWL_STYLE)
	win.SetWindowLong(hwhd, win.GWL_STYLE, currstyle & ^win.WS_SIZEBOX)
	mainWindow.Run()
	defer mainWindow.Close()
}
func AnalyisText(hrefs []string, c chan string, i int, re string, section string, book *Book) {
	filep, _ := os.Create("tmp/" + strconv.Itoa(i) + ".txt")
	defer filep.Close()
	var textresult string = ""
	fmt.Println("Start Get")
	res, err := http.Get(hrefs[i-1])
	if err != nil {
		fmt.Println("err:", err)
	}
	if res.StatusCode != 200 {
		fmt.Println("Not 200 code")
		c <- "第" + strconv.Itoa(i) + "章爬取失败"
	}
	fmt.Println("200 code")
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("err:", err)
		textresult = ""
		return
	}
	splitc := "\n \n \n"
	chapterNode := doc.Find(section)
	findCPName := regexp.MustCompile(re)
	chapter, reerr := chapterNode.Html()
	if reerr != nil {
		fmt.Println("reerr:", reerr)
		return
	}
	realchapter := findCPName.FindAllStringSubmatch(chapter, 1)[0]
	chaptername := doc.Find("h1").Text()
	article := realchapter[1]
	textresult = textresult + chaptername + splitc + article
	if book.baseburl != "www.xbiquge.la" {
		textresult = ConvertStringToUTF(textresult, "gbk", "utf-8")
	}
	filep.WriteString(textresult)
	time.Sleep(time.Second)
	c <- "第" + strconv.Itoa(i) + "章爬取成功"
	defer res.Body.Close()
}
func (book *Book) GetNewBook(firsturl string) {
	rep := regexp.MustCompile(`https://(.*?)/(.*)/`)
	data := rep.FindAllStringSubmatch(firsturl, 1)[0]
	book.baseburl = data[1]
	book.booknumber = data[2]
}
func (book *Book) GetAllChapter() ([]string, int, string) {

	res, err := http.Get("https://" + book.baseburl + "/" + book.booknumber + "/")
	fmt.Println("Get:", book.baseburl+"/"+book.booknumber+"/")
	if err != nil {
		fmt.Println("err:", err)
	}
	if res.StatusCode != 200 {
		fmt.Println("Not 200 code")
	}
	fmt.Println("200 code")
	docp, _ := goquery.NewDocumentFromReader(res.Body)
	listnode := docp.Find("dd")
	length := listnode.Length()
	if length == 0 {
		listnode = docp.Find("li")
		length = listnode.Length()
	}
	result := make([]string, length)
	bookname := ""
	if book.baseburl != "www.xbiquge.la" {
		bookname = ConvertStringToUTF(docp.Find("h1").Text(), "gbk", "utf-8") + ".txt"
	} else {
		bookname = docp.Find("h1").Text() + ".txt"
	}
	listnode.Each(func(i int, s *goquery.Selection) {
		result[i], _ = s.Children().Attr("href")
		if book.baseburl != "www.biqugesk.org" {
			result[i] = "https://" + book.baseburl + result[i]
		}
	})
	return result, length, bookname

}
func Start() {
	outTE.SetText("Start")
	CreateDir("./tmp")
	CreateDir("./dist")
	chapterc := make(chan string)
	baseurl := inTE.Text()
	var mybook Book
	mybook.GetNewBook(baseurl)
	hrefs, len, bookname := mybook.GetAllChapter()
	for i := 1; i <= len; i++ {
		go AnalyisText(hrefs, chapterc, i, remap[mybook.baseburl], docmap[mybook.baseburl], &mybook)
	}
	for i := 1; i <= len; i++ {
		fmt.Println(<-chapterc)
	}
	outTE.SetText("爬取完毕，开始整合\n")
	AppendFile("./dist/"+bookname, len)
}
