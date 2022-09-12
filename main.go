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
	baseburl      string
	booknumber    string
	chapternumber string
}
type Config struct {
	BaseurlC    string
	Lastchapter int
}

func main() {
	var baseurl string
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
		}
		configdata, _ := json.Marshal(config)
		OpenFileAndWrite(configdata, "./config.json")
	}
	var mybook Book
	fmt.Println("base", baseurl)
	mybook.GetNewBook(baseurl)
	hrefs, len, bookname := mybook.GetAllChapter()
	for i := config.Lastchapter; i <= len; i++ {
		go AnalyisText(hrefs, chapterc, i, &config)
	}
	for i := config.Lastchapter; i <= len; i++ {
		fmt.Println(<-chapterc)
	}
	fmt.Println("开始整合")
	AppendFile("./dist/"+bookname, len)
}
func AnalyisText(hrefs []string, c chan string, i int, conf *Config) {
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
	findCPName := regexp.MustCompile(`<script>([\s\S]*?)</div>([\s\S]*?)<script>`)
	chapter, reerr := chapterNode.Html()
	if reerr != nil {
		fmt.Println("reerr:", reerr)
		return
	}
	realchapter := findCPName.FindAllStringSubmatch(chapter, -1)[0]
	chaptername := doc.Find("h1").Text()
	article := realchapter[2]
	textresult = textresult + chaptername + splitc + article
	textresult = ConvertStringToUTF(textresult, "gbk", "utf-8")
	filep.WriteString(textresult)
	time.Sleep(time.Second)
	c <- "第" + strconv.Itoa(i) + "章爬取成功"
	defer res.Body.Close()
}
func (book *Book) GetNewBook(firsturl string) {
	rep := regexp.MustCompile(`https://(.*?)/(.*?)/(.*?).html`)
	data := rep.FindAllStringSubmatch(firsturl, 1)[0]
	book.baseburl = "https://" + data[1]
	book.booknumber = data[2]
	book.chapternumber = data[3]
}
func (book *Book) ShowBook() {
	fmt.Println(book.chapternumber)
}
func (book *Book) GetAllChapter() ([]string, int, string) {

	res, err := http.Get(book.baseburl + "/" + book.booknumber + "/")
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
	result := make([]string, length)
	bookname := ConvertStringToUTF(docp.Find("h1").Text(), "gbk", "utf-8") + ".txt"
	listnode.Each(func(i int, s *goquery.Selection) {
		result[i], _ = s.Children().Attr("href")
		fmt.Println(result[i])
		result[i] = book.baseburl + result[i]
	})
	return result, length, bookname

}
