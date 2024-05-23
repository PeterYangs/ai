package main

import (
	"context"
	"fmt"
	"github.com/PeterYangs/tools/file"
	"github.com/google/generative-ai-go/genai"
	"google.golang.org/api/option"
	"log"
	"strings"
)

func main() {

	ctx := context.Background()
	// Access your API key as an environment variable (see "Set up your API key" above)
	client, err := genai.NewClient(ctx, option.WithAPIKey("AIzaSyDD_bFugU_TFETT-HaC40SAB2PChzmC_dc"))
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// For text-only input, use the gemini-pro model
	model := client.GenerativeModel("gemini-pro")

	//question := "以{newtitle}为中心，按照下面的要求写一篇500字左右的中文文章！1、介绍文章{newtitle}，引出读者的兴趣，并给读者提供背景信息。2、请从多个方面对{newtitle}做详细的阐述，每个方面都要有多个自然段，并且这多个方面的小标题，字数能够控制在10汉字左右；每个自然段至少100个汉字，小标题用<h2></h2>标签包裹。\n以{newtitle}为题写一篇500字左右中文文章，文章要求：1、引人入胜：可以使用强烈的词汇或奇特的概念，让读者感到好奇。2、反映主题：应该与标题的主题紧密相关，能够准确地反映文章的主旨，让读者知道他们将要阅读什么样的内容。3、保证文章能增加搜索引擎的可见度，吸引更多的读者。"
	question := "您现在是一名优秀的文案创作大师，我需要围绕{keyword}写一篇有关区块链钱包下载应用的使用说明，先取一个小标题用<h2></h2>标签包裹放在内容的最上面，字数200字左右，请使用中文。"

	//file.ReadLine("keyword.txt", func(line []byte) {

	l := "imToken有网页版"
	//l := string(line)

	question = strings.Replace(question, "{keyword}", l, -1)

	resp, err := model.GenerateContent(ctx, genai.Text(question))
	if err != nil {
		//log.Fatal(err)

		fmt.Println(err)

		return
	}

	for _, candidate := range resp.Candidates {

		a := fmt.Sprintln(candidate.Content.Parts[0])

		file.Write(l+".txt", []byte(a))

	}

	//})

}
