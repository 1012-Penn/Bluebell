package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/brianvoe/gofakeit/v6"
)

type Post struct {
	AuthorID    int64  `json:"author_id"`
	CommunityID int64  `json:"community_id"`
	Status      int32  `json:"status"`
	Title       string `json:"title"`
	Content     string `json:"content"`
}

// GenerateFakePosts 自动生成假帖子
func GenerateFakePosts(n int, url, token string) {
	gofakeit.Seed(time.Now().UnixNano())
	client := &http.Client{}
	for i := 0; i < n; i++ {
		post := Post{
			AuthorID:    int64(gofakeit.Number(1, 100)),
			CommunityID: int64(gofakeit.Number(1, 4)),
			Status:      1,
			Title:       gofakeit.Sentence(6),
			Content:     gofakeit.Paragraph(1, 3, 12, " "),
		}
		b, _ := json.Marshal(post)
		req, _ := http.NewRequest("POST", url, bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+token)
		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("第%d条请求失败: %v\n", i+1, err)
			continue
		}
		resp.Body.Close()
		if (i+1)%100 == 0 {
			fmt.Printf("已生成%d条假帖子\n", i+1)
		}
	}
	fmt.Println("假帖子生成完毕！")
}

func main() {
	GenerateFakePosts(1000, "http://127.0.0.1:8084/api/v1/post", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2NDMxMTQ1NjA1OTI1NDc4NDAsInVzZXJuYW1lIjoidXNlcm5hbWUiLCJleHAiOjE3Nzg0MjgwODgsImlzcyI6ImJsdWViZWxsIn0.1DKcxPpB9MWVW6yOgzE6VXl8ZJnCwTrD5mctSxT1hJQ")
}
