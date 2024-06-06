package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"geminiDemo/common"
	"github.com/PeterYangs/tools"
	"github.com/PeterYangs/tools/file"
	"github.com/PuerkitoBio/goquery"
	aq "github.com/emirpasic/gods/queues/linkedlistqueue"
	"github.com/go-resty/resty/v2"
	"github.com/spf13/cast"
	"net/url"
	"os"
	"os/signal"
	"regexp"
	"strings"
	"sync"
	"syscall"
	"time"
	"unicode/utf8"
)

type Client struct {
	queue  *aq.Queue
	cxt    context.Context
	cancel context.CancelFunc
	config sync.Map
	client *resty.Client
	//ch     chan int
	wait sync.WaitGroup
}

func main() {

	client := NewClient(context.Background())

	client.Start()

}

func NewClient(cxt context.Context) *Client {

	c, cancel := context.WithCancel(cxt)

	client := resty.New()

	client.SetTimeout(2 * time.Minute)

	return &Client{cxt: c, cancel: cancel, config: sync.Map{}, client: client}
}

func (c *Client) Start() {

	sigs := make(chan os.Signal, 1)

	//退出信号
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	quit := make(chan bool)

	go func() {

		<-sigs

		c.cancel()

		fmt.Println("退出")

		quit <- true

	}()

	c.loadConfig()

	num, _ := c.getConfig("线程数")

	//c.ch = make(chan int, cast.ToInt(num))

	c.wait = sync.WaitGroup{}

	common.CopyFile("关键词.txt", "关键词-备份.txt")

	fmt.Println("正在加载关键词。。。")

	all, _ := file.Read("关键词.txt")

	queue := aq.New()

	arr := tools.Explode("\n", string(all))

	for _, s := range arr {

		queue.Enqueue(s)

	}

	fmt.Println("关键词加载完毕")

	c.queue = queue

	for i := 0; i < cast.ToInt(num); i++ {

		c.wait.Add(1)

		go c.deal()

	}

	c.wait.Wait()

	<-quit

	fff, _ := os.OpenFile("关键词.txt", os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)

	defer fff.Close()

	for {

		k, ok := c.queue.Dequeue()

		if !ok {

			break
		}

		fff.Write([]byte(k.(string) + "\n"))

	}

	fmt.Println("安全退出")

}

func (c *Client) deal() {

	defer func() {

		c.wait.Done()

	}()

	for {

		select {

		case <-c.cxt.Done():

			return

		default:

			k, ok := c.queue.Dequeue()

			if !ok {

				c.cancel()

				return
			}

			keyword := strings.TrimSpace(k.(string))

			time.Sleep(1 * time.Second)

			cmdCountStr, cmdCountOk := c.getConfig("小标题数量")

			if !cmdCountOk {

				c.cancel()

			}

			fileName := keyword

			cmdCountArr := tools.Explode("-", cmdCountStr)

			cmdCount := tools.MtRand(cast.ToInt64(cmdCountArr[0]), cast.ToInt64(cmdCountArr[1]))

			baiduNewTitle, baiduNewTitleErr := c.getBaiduKeyword(keyword)

			if baiduNewTitleErr != nil {

				newTitleCmd, newTitleCmdOk := c.getConfig("伪原创标题的指令")

				if newTitleCmdOk {

					newTitle, newTitleCmdErr := c.requestAi(strings.Replace(newTitleCmd, "{keyword}", keyword, -1))

					if newTitleCmdErr == nil {

						//
						fileName = c.mergeTitle(newTitle, fileName)

					} else {

						fmt.Println("原创标题获取失败", newTitleCmdErr)

					}

				}

			} else {

				//
				//fileName = strings.TrimSpace(baiduNewTitle) + "-" + fileName
				fileName = c.mergeTitle(strings.TrimSpace(baiduNewTitle), fileName)

			}

			fileName += ".txt"

			//ff, _ := os.OpenFile("回答/原版-"+fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)

			content := ""

			isOk := true

		DD:
			for i := 0; i < int(cmdCount); i++ {

				select {

				case <-c.cxt.Done():

					break DD

				default:

					cmd, ook := c.getConfig("第" + cast.ToString(i+1) + "段文章回答命令")

					if !ook {

						continue
					}

					cmd = strings.Replace(cmd, "{keyword}", keyword, -1)

					fmt.Println("正在生成", keyword, "第"+cast.ToString(i+1)+"段", cmd)

					str, err := c.requestAi(cmd)

					if err != nil {

						fmt.Println(keyword, "生成错误("+cmd+")", err)

						isOk = false

						break DD

					}

					if strings.Contains(str, "<body>") && strings.Contains(str, "</body>") {

						str = strings.Replace(str, "<h1>", "<h3>", -1)
						str = strings.Replace(str, "</h1>", "</h3>", -1)

						str = strings.Replace(str, "<h2>", "<h3>", -1)
						str = strings.Replace(str, "</h2>", "</h3>", -1)

						str = strings.Replace(str, "<h4>", "<h3>", -1)
						str = strings.Replace(str, "</h4>", "</h3>", -1)

						doc, _ := goquery.NewDocumentFromReader(strings.NewReader(str))

						str, _ = doc.Find("body").Html()
					}

					content += str + "\n"

				}

			}

			//fmt.Println(isOk, "-------")

			if !isOk {

				//f.Close()

				//os.Remove("回答/" + fileName)

			} else {

				select {
				case <-c.cxt.Done():

					return

				default:

				}

				f, _ := os.OpenFile("回答/"+fileName, os.O_CREATE|os.O_RDWR|os.O_TRUNC, 0644)

				contentTemp := ""

				arr := tools.Explode("\n", content)

				for _, s := range arr {

					contentTemp += c.dealAiContent(s) + "\n"

				}

				contentTemp = c.dealAiWord(contentTemp)

				f.Write([]byte(contentTemp))

				f.Close()

			}

			//ff.Close()

		}

	}

}

func (c *Client) loadConfig() {

	file.ReadLine("聚合配置.txt", func(line []byte) {

		l := string(line)

		if strings.TrimSpace(l) != "" {

			//key:=tools.SubStr()

			key := strings.TrimSpace(tools.SubStr(l, 0, common.IndexOf(l, "=")))

			value := strings.TrimSpace(tools.SubStr(l, common.IndexOf(l, "=")+1, -1))

			c.config.Store(key, value)

		}

	})

}

func (c *Client) getConfig(key string) (string, bool) {

	v, ok := c.config.Load(key)

	vv := strings.TrimSpace(v.(string))

	return vv, ok

}

func (c *Client) requestAi(question string) (string, error) {

	ai, _ := c.getConfig("ai")

	switch ai {

	case "gemini":

		return c.requestGemini(question)

	case "gpt3.5":

		return c.requestGpt35(question)

	}

	return c.requestGemini(question)

}

func (c *Client) requestGemini(question string) (string, error) {

	question = url.QueryEscape(question)

	for j := 0; j < 3; j++ {

		rsp, err := c.client.R().SetContext(c.cxt).Get("http://38.207.200.226:8199/gemini?wd=" + question + "&token=UUTuDWCe4l4jfGMwsMnU")

		if rsp.StatusCode() != 200 {

			select {
			case <-c.cxt.Done():

				return "", errors.New("安全退出")

			default:

			}

			fmt.Println("状态码异常(" + rsp.Status() + ")" + ",msg:" + rsp.String())

			continue

		}

		if err != nil {

			select {
			case <-c.cxt.Done():

				return "", errors.New("安全退出")

			default:

			}

			fmt.Println(err)

			continue
		}

		return rsp.String(), nil

	}

	return "", errors.New("错误过多，放弃")

}

func (c *Client) requestGpt35(question string) (string, error) {

	question = url.QueryEscape(question)

	for j := 0; j < 3; j++ {

		rsp, err := c.client.R().SetContext(c.cxt).Get("http://chatgpt.zf678.cn:8199/chat-gpt?wd=" + question + "&token=UUTuDWCe4l4jfGMwsMnU")

		if rsp.StatusCode() != 200 {

			select {
			case <-c.cxt.Done():

				return "", errors.New("安全退出")

			default:

			}

			fmt.Println("状态码异常(" + rsp.Status() + ")" + ",msg:" + rsp.String())

			continue

		}

		if err != nil {

			select {
			case <-c.cxt.Done():

				return "", errors.New("安全退出")

			default:

			}

			fmt.Println(err)

			continue
		}

		return rsp.String(), nil

	}

	return "", errors.New("错误过多，放弃")

}

func (c *Client) dealAiContent(str string) string {

	contentType, _ := c.getConfig("内容格式")

	switch contentType {

	case "md":

		if strings.TrimSpace(str) == "" {

			return ""
		}

		re1 := regexp.MustCompile("^### (.*)").FindStringSubmatch(str)

		if len(re1) > 0 {

			return "<h3>" + re1[1] + "</h3>"
		}

		re2 := regexp.MustCompile("^## (.*)").FindStringSubmatch(str)

		if len(re2) > 0 {

			return "<h3>" + re2[1] + "</h3>"
		}

		re3 := regexp.MustCompile(`^\*\*(.*)\*\*`).FindStringSubmatch(str)

		if len(re3) > 0 {

			return "<h3>" + re3[1] + "</h3>"
		}

		str = regexp.MustCompile(`^\s*\* `).ReplaceAllString(str, "")

		r, _ := regexp.Compile(`\*\*(.*)\*\*`)

		s := r.FindAllStringSubmatch(str, -1)

		for _, i2 := range s {

			str = strings.Replace(str, i2[0], "<strong>"+i2[1]+"</strong>", 1)

		}

		return "<p>" + str + "</p>"

	case "html":

		doc, _ := goquery.NewDocumentFromReader(strings.NewReader(str))

		text := strings.TrimSpace(doc.Text())

		if text == "总结" {

			return ""
		}

		str = strings.Replace(str, "<p><h3>", "<p>", -1)
		str = strings.Replace(str, "</h3></p>", "/<p>", -1)

	}

	return str
}

func (c *Client) dealAiWord(str string) string {

	arr := []string{"总而言之，", "总的来说，", "综上所述，", "总体来说，", "其次是", "在游戏中，", "总之，", "此外，", "更重要的是，", "展望未来，", "但是，", "值得注意的是，", "然而，", "说真的，", "唯一的小遗憾是，", "目前，", "不过，", "近年来，", "总体而言，", "最棒的是，", "综上所述，", "首先，", "其次，", "最后，", "标题：", "总结一下：", "通过以上步骤，", "另外，", "总结而言，", "值得一提的是，", "至此，"}

	for _, s := range arr {

		str = strings.Replace(str, s, "", -1)

	}

	return str

}

// 百度下拉词
func (c *Client) getBaiduKeyword(l string) (string, error) {

	escapeUrl := url.QueryEscape(l)

	u := "https://www.baidu.com/sugrec?pre=1&p=3&ie=utf-8&json=1&prod=pc&from=pc_web&wd=" + escapeUrl

	client := resty.New()

	rsp, e := client.R().Get(u)

	if e != nil {

		return "", e
	}

	type g struct {
		Q string `json:"q"`
	}

	type r struct {
		G []g `json:"g"`
	}

	var rr r

	jErr := json.Unmarshal(rsp.Body(), &rr)

	if jErr != nil {

		return "", jErr
	}

	if len(rr.G) > 0 {

		return rr.G[tools.MtRand(0, int64(len(rr.G)-1))].Q, nil

	}

	return "", errors.New("结果为空")

}

func (c *Client) mergeTitle(t1, t2 string) string {

	d, _ := c.getConfig("双标题连接符号")

	arr := tools.Explode("|", d)

	s := arr[tools.MtRand(0, int64(len(arr)-1))]

	length := utf8.RuneCountInString(s)

	if length == 1 {

		return t1 + s + t2

	} else {

		sss := []rune(s)

		return t1 + string(sss[0]) + t2 + string(sss[1])
	}

}
