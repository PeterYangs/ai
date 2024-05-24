package main

import (
	"context"
	"fmt"
	"geminiDemo/db"
	"geminiDemo/model"
	aq "github.com/emirpasic/gods/queues/arrayqueue"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"net/http"
	"time"
)

type ChatGptServer struct {
	queue *aq.Queue
	//client *resty.Client
}

func NewChatGptServer() *ChatGptServer {

	queue := aq.New()

	//client := resty.New()

	//if os.Getenv("CHAT_PROXY") == "true" {
	//
	//	client.SetProxy("http://127.0.0.1:33210")
	//
	//}

	return &ChatGptServer{
		queue: queue,
		//
	}
}

func (c *ChatGptServer) Start() {

	var ks []model.Key

	db.GetDb().Model(&model.Key{}).Where("status = 1 and lock_time = 0 and type = 'gemini'").Find(&ks)

	//if tErr != nil {
	//
	//	panic("key已用完")
	//
	//	return
	//}

	for _, t := range ks {

		c.queue.Enqueue(t)

	}

	c.httpServer()

}

func (c *ChatGptServer) httpServer() {

	http.HandleFunc("/gemini", c.ProMax)
	//http.HandleFunc("/pp", pp)

	http.ListenAndServe(":8199", nil)

}

func (c *ChatGptServer) ProMax(w http.ResponseWriter, req *http.Request) {

	wd := req.URL.Query().Get("wd")

	token := req.URL.Query().Get("token")

	if token != "UUTuDWCe4l4jfGMwsMnU" {

		w.WriteHeader(500)

		w.Write([]byte("错误"))

		return
	}

	if wd == "" {

		w.WriteHeader(500)

		w.Write([]byte("内容为空"))

		return

	}

	t, ok := c.queue.Dequeue()

	if ok {

		realKey := t.(model.Key)

		//fmt.Println("获取key成功", realKey.Key)

		c.logic(realKey, wd, w)

	} else {

		for {

			select {

			case <-time.After(time.Second * 15):

				w.WriteHeader(500)

				w.Write([]byte("获取key等待超时"))

				return

			case <-time.After(time.Millisecond * 500):

				tt, okk := c.queue.Dequeue()

				if okk {

					realKey := tt.(model.Key)

					//fmt.Println("获取key成功", realKey.Key)

					c.logic(realKey, wd, w)

					return
				}

			}
		}

	}

}

func (c *ChatGptServer) delay(key model.Key, delay time.Duration) {

	time.Sleep(delay)

	c.queue.Enqueue(key)

}

func (c *ChatGptServer) logic(key model.Key, wd string, w http.ResponseWriter) {

	isBack := true

	defer func() {

		if isBack {

			//延迟回队列
			go c.delay(key, 10*time.Second)

		}

	}()

	//fmt.Println("key", key.Key)

	ctx, can := context.WithTimeout(context.Background(), 2*time.Minute)

	defer can()

	// Access your API key as an environment variable (see "Set up your API key" above)
	client, err := genai.NewClient(ctx, option.WithAPIKey("AIzaSyDD_bFugU_TFETT-HaC40SAB2PChzmC_dc"))
	if err != nil {
		//log.Fatal(err)

		w.WriteHeader(500)

		w.Write([]byte(err.Error()))

		fmt.Println(err)

		return

	}
	defer client.Close()

	// For text-only input, use the gemini-pro model
	mm := client.GenerativeModel("gemini-pro")

	//question := "以{newtitle}为中心，按照下面的要求写一篇500字左右的中文文章！1、介绍文章{newtitle}，引出读者的兴趣，并给读者提供背景信息。2、请从多个方面对{newtitle}做详细的阐述，每个方面都要有多个自然段，并且这多个方面的小标题，字数能够控制在10汉字左右；每个自然段至少100个汉字，小标题用<h2></h2>标签包裹。\n以{newtitle}为题写一篇500字左右中文文章，文章要求：1、引人入胜：可以使用强烈的词汇或奇特的概念，让读者感到好奇。2、反映主题：应该与标题的主题紧密相关，能够准确地反映文章的主旨，让读者知道他们将要阅读什么样的内容。3、保证文章能增加搜索引擎的可见度，吸引更多的读者。"
	//question := "您现在是一名优秀的文案创作大师，我需要围绕{keyword}写一篇有关区块链钱包下载应用的使用说明，先取一个小标题用<h2></h2>标签包裹放在内容的最上面，字数200字左右，请使用中文。"
	question := wd

	//file.ReadLine("keyword.txt", func(line []byte) {

	//l := "imToken有网页版"
	////l := string(line)
	//
	//question = strings.Replace(question, "{keyword}", l, -1)

	resp, err := mm.GenerateContent(ctx, genai.Text(question))
	if err != nil {
		//log.Fatal(err)

		//fmt.Println(err)

		w.WriteHeader(500)

		w.Write([]byte(err.Error()))

		return
	}

	content := ""

	for _, candidate := range resp.Candidates {

		a := fmt.Sprintln(candidate.Content.Parts[0])

		//file.Write(l+".txt", []byte(a))

		content += a

	}

	w.Write([]byte(content))

	//rsp, err := c.client.R().SetHeaders(map[string]string{
	//	"Authorization": "Bearer " + token.Token,
	//	"Content-Type":  "application/json",
	//}).SetBody(map[string]interface{}{
	//	"model": "gpt-3.5-turbo-16k",
	//	"messages": []interface{}{
	//		map[string]interface{}{"role": "user", "content": wd},
	//	},
	//}).Post("https://api.openai.com/v1/chat/completions")
	//
	////fmt.Println(token.Token)
	////time.Sleep(3 * time.Second)
	//
	//if err != nil {
	//
	//	fmt.Println("响应错误：", rsp.String())
	//
	//	w.WriteHeader(500)
	//
	//	w.Write([]byte(err.Error()))
	//
	//	return
	//}
	//
	//if rsp.StatusCode() != 200 {
	//
	//	fmt.Println("状态码不等于200", rsp.String())
	//
	//	w.WriteHeader(500)
	//
	//	if strings.Contains(rsp.String(), "insufficient_quota") {
	//
	//		isBack = false
	//
	//		db.GetDb().Model(&model.Tokens{}).Where("id = ?", token.Id).Where("status = 1").Updates(map[string]interface{}{"status": 2})
	//
	//	} else if strings.Contains(rsp.String(), "The OpenAI account associated with this API key has been deactivated") {
	//
	//		//账号被禁
	//		isBack = false
	//
	//		db.GetDb().Model(&model.Tokens{}).Where("id = ?", token.Id).Where("status = 1").Updates(map[string]interface{}{"status": 3})
	//
	//	}
	//
	//	w.Write([]byte(rsp.String()))
	//
	//	return
	//
	//}
	//
	//type message struct {
	//	Role    string `json:"role"`
	//	Content string `json:"content"`
	//	Index   int    `json:"index"`
	//}
	//
	//type text struct {
	//	Message message `json:"message"`
	//}
	//
	//type errors struct {
	//	Message string `json:"message"`
	//	Type    string `json:"type"`
	//	Code    string `json:"code"`
	//}
	//
	//type res struct {
	//	Id      string `json:"id"`
	//	Model   string `json:"model"`
	//	Choices []text
	//	Error   errors `json:"error"`
	//}
	//
	//var r res
	//
	//jErr := json.Unmarshal(rsp.Body(), &r)
	//
	//if jErr != nil {
	//
	//	fmt.Println(jErr)
	//
	//	w.WriteHeader(500)
	//
	//	w.Write([]byte(jErr.Error()))
	//
	//	return
	//}
	//
	//fmt.Println(r)
	//
	//if r.Error.Message != "" {
	//
	//	w.WriteHeader(500)
	//
	//	w.Write([]byte(r.Error.Message))
	//
	//	//帐号过期置为2
	//	if r.Error.Type == "insufficient_quota" {
	//
	//		isBack = false
	//
	//		db.GetDb().Model(&model.Tokens{}).Where("id = ?", token.Id).Where("status = 1").Updates(map[string]interface{}{"status": 2})
	//
	//	}
	//
	//	//账号被禁用
	//	if r.Error.Code == "account_deactivated" {
	//
	//		isBack = false
	//
	//		db.GetDb().Model(&model.Tokens{}).Where("id = ?", token.Id).Where("status = 1").Updates(map[string]interface{}{"status": 3})
	//	}
	//
	//	return
	//}
	//
	//if len(r.Choices) <= 0 {
	//
	//	fmt.Println(rsp.String())
	//
	//	w.WriteHeader(500)
	//
	//	w.Write([]byte("返回内容为空"))
	//
	//	return
	//}
	//
	//w.WriteHeader(200)
	//
	//w.Write([]byte(r.Choices[0].Message.Content))

}

func init() {

	err := godotenv.Load(".env")

	if err != nil {
		panic("配置文件加载失败")
	}

}

func main() {

	c := NewChatGptServer()

	c.Start()

}
