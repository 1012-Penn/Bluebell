package queue

import (
	"bluebell/dao/mysql"
	"context"
	"log"
	"sync"
	"time"
)

// VoteMessage 投票消息结构
type VoteMessage struct {
	PostID    int64 `json:"post_id"`
	UserID    int64 `json:"user_id"`
	VoteValue int8  `json:"vote_value"`
	Timestamp int64 `json:"timestamp"`
}

// VoteQueue 投票消息队列
type VoteQueue struct {
	messages chan VoteMessage
	wg       sync.WaitGroup
	ctx      context.Context
	cancel   context.CancelFunc
}

var (
	voteQueue *VoteQueue
	once      sync.Once
)

// InitVoteQueue 初始化投票队列
func InitVoteQueue() {
	once.Do(func() {
		ctx, cancel := context.WithCancel(context.Background())
		voteQueue = &VoteQueue{
			messages: make(chan VoteMessage, 10000), // 缓冲区1万条消息
			ctx:      ctx,
			cancel:   cancel,
		}
		voteQueue.startWorker()
	})
}

// EnqueueVote 入队投票消息
func EnqueueVote(postID, userID int64, voteValue int8) {
	if voteQueue != nil {
		msg := VoteMessage{
			PostID:    postID,
			UserID:    userID,
			VoteValue: voteValue,
			Timestamp: time.Now().Unix(),
		}
		select {
		case voteQueue.messages <- msg:
			// 消息入队成功
		default:
			log.Printf("投票队列已满，丢弃消息: %+v", msg)
		}
	}
}

// startWorker 启动工作协程
func (vq *VoteQueue) startWorker() {
	vq.wg.Add(1)
	go func() {
		defer vq.wg.Done()
		ticker := time.NewTicker(5 * time.Second) // 每5秒批量处理
		defer ticker.Stop()

		var batch []VoteMessage

		for {
			select {
			case msg := <-vq.messages:
				batch = append(batch, msg)
				if len(batch) >= 100 { // 批量处理100条
					vq.processBatch(batch)
					batch = batch[:0]
				}
			case <-ticker.C:
				if len(batch) > 0 {
					vq.processBatch(batch)
					batch = batch[:0]
				}
			case <-vq.ctx.Done():
				if len(batch) > 0 {
					vq.processBatch(batch)
				}
				return
			}
		}
	}()
}

// processBatch 批量处理投票消息
func (vq *VoteQueue) processBatch(batch []VoteMessage) {
	if len(batch) == 0 {
		return
	}

	// 批量写入MySQL
	for _, msg := range batch {
		err := mysql.SaveVoteData(msg.PostID, msg.UserID, msg.VoteValue)
		if err != nil {
			log.Printf("批量写入投票数据失败: %v, 消息: %+v", err, msg)
		}
	}

	log.Printf("批量处理了 %d 条投票消息", len(batch))
}

// Close 关闭队列
func (vq *VoteQueue) Close() {
	if vq.cancel != nil {
		vq.cancel()
	}
	vq.wg.Wait()
}
