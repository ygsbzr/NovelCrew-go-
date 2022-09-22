package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
)

func HasDir(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
func CreateDir(path string) {
	exists, err := HasDir(path)
	if err != nil {
		fmt.Println("文件夹异常:", err)
		return
	}
	if exists {
		fmt.Println("文件夹已存在")
	} else {
		errm := os.Mkdir(path, os.ModePerm)
		if errm != nil {
			fmt.Println("err Create:", errm)
			return
		} else {
			fmt.Println("Successful mkdir")
		}
	}
}
func AppendFile(distpath string, cnum int) {
	dist, err := os.OpenFile(distpath, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0644)
	outtestr := outTE.Text()
	if err != nil {
		return
	}
	defer dist.Close()
	for i := 1; i <= cnum; i++ {

		pagefile, pageerr := os.Open("tmp/" + strconv.Itoa(i) + ".txt")
		if pageerr != nil {
			return
		}
		read := bufio.NewReader(pagefile)
		result := ""
		chaptername := true
		for {
			line, _, readerr := read.ReadLine()
			result += string(line)
			if chaptername {
				result += "\n\n\n"
				chaptername = false
			}
			if readerr != nil {
				if readerr == io.EOF {
					result = strings.ReplaceAll(result, "<br/>", "\n")
					result = strings.ReplaceAll(result, "聽", "")
					result = strings.ReplaceAll(result, "&nbsp;", " ")
					result = strings.ReplaceAll(result, "<br>", "\n")
					result = strings.ReplaceAll(result, "</a>", "")
					result = strings.ReplaceAll(result, "<p>", "")
					result = strings.ReplaceAll(result, "</p>", "\n")
					dist.WriteString(result + "\n \n \n \n \n")
					break
				}

			}
		}
		pagefile.Close()
	}
	outtestr += "整理完毕:" + strings.Replace(distpath, "./dist/", "", 1) + " " + "共" + strconv.Itoa(cnum) + "章\n"
	outTE.SetText(outtestr)
	os.RemoveAll("./tmp")
}
func OpenFileAndRead(path string) []byte {
	f, err := os.Open(path)
	if err != nil {
		return []byte{'\n'}
	}
	defer f.Close()
	buf := make([]byte, 1024*2)
	n, _ := f.Read(buf)
	return buf[:n]
}
func OpenFileAndWrite(buf []byte, path string) {
	f, err := os.OpenFile(path, os.O_WRONLY|os.O_CREATE, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	f.Write(buf)
}
