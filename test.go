package main

import (
	"fmt"
	"github.com/PeterYangs/tools"
	"regexp"
	"strings"
)

func main() {

	str := `## 《王者荣耀》：一款风靡全球的MOBA手游巨作

### 游戏玩法

《王者荣耀》是一款5v5多人在线竞技游戏（MOBA），玩家扮演英雄角色，与队友合作，在战场上与敌方英雄和防御塔战斗。游戏目标是摧毁敌方基地（水晶），同时保护己方水晶免受摧毁。

玩家可以从一个庞大的英雄库中选择英雄，每个英雄都有独特的技能和能力。团队合作和策略在《王者荣耀》中至关重要，玩家需要协调策略、沟通信息并共同努力才能取得胜利。

游戏玩法紧张刺激，具有以下特点：

* **实时对战：**与来自世界各地的真人玩家进行实时对战。
* **多种游戏模式：**包括经典的5v5对战、排位赛和休闲模式。
* **各种英雄：**拥有超过100个英雄，每个英雄都有独特的技能和能力。
* **地图设计：**游戏地图设计精良，具有多个战场，包括丛林、河流和据点。

### 游戏特色

除了激动人心的游戏玩法外，《王者荣耀》还拥有以下显着的特色：

* **精致的图形：**游戏画面精美逼真，角色形象和场景设计细致入微。
* **英雄配音：**每个英雄都有专业声优配音，为游戏增添了沉浸感和个性。
* **公平竞技：**游戏采用匹配机制，确保玩家在相对公平的环境中竞争。
* **社交互动：**玩家可以加入战队，与朋友或其他玩家组队，分享游戏体验和策略。
* **持续更新：**游戏定期更新，添加新英雄、新皮肤和新游戏模式，保持游戏的新鲜感。

**游戏特色**

《王者荣耀》是一款制作精良、内容丰富的MOBA手游巨作。其紧张刺激的游戏玩法、庞大的英雄库和精致的画面，为玩家带来了前所未有的游戏体验。此外，游戏的社交互动功能和持续更新内容，确保了其长期的的可玩性和吸引力。

对于MOBA爱好者和手游玩家来说，强烈推荐《王者荣耀》。其精美的画面、多样化的英雄和激烈的对战，将为您带来无与伦比的游戏乐趣。`

	arr := tools.Explode("\n", str)

	for _, s := range arr {

		fmt.Println(deal(s))

	}

}

func deal(str string) string {

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

	str = regexp.MustCompile(`^\* `).ReplaceAllString(str, "")

	r, _ := regexp.Compile(`\*\*(.*)\*\*`)

	s := r.FindAllStringSubmatch(str, -1)

	for _, i2 := range s {

		str = strings.Replace(str, i2[0], "<strong>"+i2[1]+"</strong>", 1)

	}

	return "<p>" + str + "</p>"
}
