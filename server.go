package main

import (
	"context"
	"fmt"
	"geminiDemo/db"
	"geminiDemo/model"
	aq "github.com/emirpasic/gods/queues/linkedlistqueue"
	"github.com/google/generative-ai-go/genai"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
	"net/http"
	"strings"
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

			case <-time.After(time.Minute * 1):

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

	no := false

	defer func() {

		if !no {

			if isBack {

				//延迟回队列
				go c.delay(key, 1*time.Minute)

			} else {

				go c.delay(key, 6*time.Hour)
			}
		}
	}()

	//fmt.Println("key", key.Key)

	ctx, can := context.WithTimeout(context.Background(), 2*time.Minute)

	defer can()

	// Access your API key as an environment variable (see "Set up your API key" above)
	client, err := genai.NewClient(ctx, option.WithAPIKey(key.Key))
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

	question := wd

	resp, err := mm.GenerateContent(ctx, genai.Text(question))
	if err != nil {

		if strings.Contains(strings.ToLower(err.Error()), strings.ToLower("googleapi: Error 429: Resource has been exhausted (e.g. check quota)")) {

			isBack = false

		}

		if strings.Contains(strings.ToLower(err.Error()), strings.ToLower("googleapi: Error 403: Permission denied: Consumer")) {

			db.GetDb().Model(&model.Key{}).Where("id = ?", key.Id).Update("status", 2)

			no = true
		}

		if strings.Contains(strings.ToLower(err.Error()), strings.ToLower("CONSUMER_SUSPENDED domain = googleapis.com metadata")) {

			db.GetDb().Model(&model.Key{}).Where("id = ?", key.Id).Update("status", 2)

			no = true
		}

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

	if strings.Contains(strings.ToLower(content), strings.ToLower("googleapi: Error 429: Resource has been exhausted (e.g. check quota)")) {

		isBack = false

	}

	if strings.Contains(strings.ToLower(content), strings.ToLower("googleapi: Error 403: Permission denied: Consumer")) {

		db.GetDb().Model(&model.Key{}).Where("id = ?", key.Id).Update("status", 2)

		no = true
	}

	if strings.Contains(strings.ToLower(content), strings.ToLower("CONSUMER_SUSPENDED domain = googleapis.com metadata")) {

		db.GetDb().Model(&model.Key{}).Where("id = ?", key.Id).Update("status", 2)

		no = true
	}

	if strings.Contains(strings.ToLower(content), strings.ToLower("googleapi: Error")) {

		w.WriteHeader(500)

		w.Write([]byte(content))

		return
	}

	w.Write([]byte(content))

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
