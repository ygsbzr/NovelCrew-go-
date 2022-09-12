package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"regexp"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

type Book struct {
	baseburl   string
	booknumber string
}
type Config struct {
	BaseurlC    string
	Lastchapter int
	BookWebUrl  string
}

func main() {
	var baseurl string
	var remap map[string]string = make(map[string]string)
	remap["www.zhhbq.com"] = `<script>[\s\S]*?</div>([\s\S]*?)<script>`
	remap["www.xbiquge.la"] = `([\s\S]*?)<p>`
	remap["www.ddyueshu.com"] = `([\s\S]*?)<p>`
	remap["www.qu-la.com"] = `([\s\S]*)`

	var config Config
	CreateDir("./tmp")
	CreateDir("./dist")
	chapterc := make(chan string)
	if hasconf, _ := HasDir("./config.json"); hasconf {
		json.Unmarshal(OpenFileAndRead("./config.json"), &config)
		baseurl = config.BaseurlC
	} else {
		fmt.Println("无config,请输入基础网址")
		fmt.Scan(&baseurl)
		config = Config{
			BaseurlC:    baseurl,
			Lastchapter: 1,
			BookWebUrl:  "",
		}
		configdata, _ := json.Marshal(config)
		OpenFileAndWrite(configdata, "./config.json")
	}
	var mybook Book
	mybook.GetNewBook(baseurl)
	config.BookWebUrl = mybook.baseburl
	hrefs, len, bookname := mybook.GetAllChapter()
	for i := config.Lastchapter; i <= len; i++ {
		go AnalyisText(hrefs, chapterc, i, &config, remap[config.BookWebUrl])
	}
	for i := config.Lastchapter; i <= len; i++ {
		fmt.Println(<-chapterc)
	}
	fmt.Println("开始整合")
	AppendFile("./dist/"+bookname, len)
}
func AnalyisText(hrefs []string, c chan string, i int, conf *Config, re string) {
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
		conf.Lastchapter = i
		return
	}
	fmt.Println("200 code")
	doc, err := goquery.NewDocumentFromReader(res.Body)
	if err != nil {
		fmt.Println("err:", err)
		textresult = ""
		return
	}
	splitc := "\n \n \n"
	chapterNode := doc.Find("div#content")
	if chapterNode.Length() == 0 {
		chapterNode = doc.Find("div[id=chapter-title]~div")
	}
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
	textresult = ConvertStringToUTF(textresult, "gbk", "utf-8")
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
	fmt.Println("len:", length)
	if length == 0 {
		listnode = docp.Find("li")
		length = listnode.Length()
		fmt.Println("len:", length)
	}
	result := make([]string, length)
	bookname := ConvertStringToUTF(docp.Find("h1").Text(), "gbk", "utf-8") + ".txt"
	listnode.Each(func(i int, s *goquery.Selection) {
		result[i], _ = s.Children().Attr("href")
		result[i] = "https://" + book.baseburl + result[i]
		fmt.Println(result[i])
	})
	return result, length, bookname

}
