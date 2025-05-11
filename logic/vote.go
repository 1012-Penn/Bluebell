package logic

import (
	"bluebell/dao/mysql"
	"bluebell/models"
	"fmt"
	"sync"

	"context"

	"github.com/segmentio/kafka-go"
)

// 推荐阅读
// 基于用户投票的相关算法：http://www.ruanyifeng.com/blog/algorithm/

// 本项目使用简化版的投票分数
// 投一票就加432分   86400/200  --> 200张赞成票可以给你的帖子续一天

/* 投票的几种情况：
direction=1时，有两种情况：
	1. 之前没有投过票，现在投赞成票    --> 更新分数和投票记录
	2. 之前投反对票，现在改投赞成票    --> 更新分数和投票记录
direction=0时，有两种情况：
	1. 之前投过赞成票，现在要取消投票  --> 更新分数和投票记录
	2. 之前投过反对票，现在要取消投票  --> 更新分数和投票记录
direction=-1时，有两种情况：
	1. 之前没有投过票，现在投反对票    --> 更新分数和投票记录
	2. 之前投赞成票，现在改投反对票    --> 更新分数和投票记录

投票的限制：
每个贴子自发表之日起一个星期之内允许用户投票，超过一个星期就不允许再投票了。
	1. 到期之后将redis中保存的赞成票数及反对票数存储到mysql表中
	2. 到期之后删除那个 KeyPostVotedZSetPF
*/

var likeCache = make(map[string]int64) // 本地缓存，key为postID
var likeCacheLock sync.Mutex
var kafkaWriter *kafka.Writer

func InitKafkaWriter(brokers []string, topic string) {
	kafkaWriter = &kafka.Writer{
		Addr:  kafka.TCP(brokers...),
		Topic: topic,
	}
}

// VoteForPost 为帖子投票的函数
func VoteForPost(userID int64, p *models.ParamVoteData) error {
	likeCacheLock.Lock()
	likeCache[p.PostID]++
	likeCacheLock.Unlock()

	// 发送消息到Kafka
	msg := kafka.Message{
		Key:   []byte(p.PostID),
		Value: []byte(fmt.Sprintf("%d,%s,%d", userID, p.PostID, p.Direction)),
	}
	if kafkaWriter != nil {
		_ = kafkaWriter.WriteMessages(context.Background(), msg)
	}
	return nil // 不再直接写MySQL和Redis
}

// 启动Kafka消费者，消费点赞消息并写入MySQL
func StartVoteConsumer(brokers []string, topic string) {
	r := kafka.NewReader(kafka.ReaderConfig{
		Brokers: brokers,
		Topic:   topic,
		GroupID: "vote_group",
	})
	go func() {
		for {
			m, err := r.ReadMessage(context.Background())
			if err != nil {
				continue
			}
			// 解析消息内容
			var userID int64
			var postID string
			var direction int8
			fmt.Sscanf(string(m.Value), "%d,%s,%d", &userID, &postID, &direction)
			// 写入MySQL
			if direction == 0 {
				_ = mysql.DeletePostVote(userID, postID)
			} else {
				_ = mysql.InsertPostVote(userID, postID, direction)
			}
		}
	}()
}
