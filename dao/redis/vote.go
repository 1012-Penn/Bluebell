package redis

import (
	"context"
	"errors"
	"math"
	"strconv"
	"time"

	"github.com/go-redis/redis/v8"
)

// 推荐阅读
// 基于用户投票的相关算法：http://www.ruanyifeng.com/blog/algorithm/

// 本项目使用简化版的投票分数
// 投一票就加432分   86400/200  --> 200张赞成票可以给你的帖子续一天

/* 投票的几种情况：
   direction=1时，有两种情况：
   	1. 之前没有投过票，现在投赞成票    --> 更新分数和投票记录  差值的绝对值：1  +432
   	2. 之前投反对票，现在改投赞成票    --> 更新分数和投票记录  差值的绝对值：2  +432*2
   direction=0时，有两种情况：
   	1. 之前投过反对票，现在要取消投票  --> 更新分数和投票记录  差值的绝对值：1  +432
	2. 之前投过赞成票，现在要取消投票  --> 更新分数和投票记录  差值的绝对值：1  -432
   direction=-1时，有两种情况：
   	1. 之前没有投过票，现在投反对票    --> 更新分数和投票记录  差值的绝对值：1  -432
   	2. 之前投赞成票，现在改投反对票    --> 更新分数和投票记录  差值的绝对值：2  -432*2

   投票的限制：
   每个贴子自发表之日起一个星期之内允许用户投票，超过一个星期就不允许再投票了。
   	1. 到期之后将redis中保存的赞成票数及反对票数存储到mysql表中
   	2. 到期之后删除那个 KeyPostVotedZSetPF
*/

// 实际生产环境下 context.Background() 按需替换

const (
	// oneWeekInSeconds: 一周的秒数
	// 命名逻辑：one + Week + In + Seconds（一周的秒数）
	// 7天 * 24小时 * 3600秒 = 604800秒
	oneWeekInSeconds = 7 * 24 * 3600

	// scorePerVote: 每票的分数值
	// 命名逻辑：score + Per + Vote（每次投票的分数）
	// 432分 = 86400秒/200票，即200张赞成票可以让帖子在热门榜上多待一天
	scorePerVote = 432 // 每一票值多少分
)

var (
	// ErrVoteTimeExpire: 投票时间过期错误
	// 命名逻辑：Err + Vote + Time + Expire（投票时间过期错误）
	// 用于表示帖子发布超过一周，不允许再投票
	ErrVoteTimeExpire = errors.New("投票时间已过")

	// ErrVoteRepeated: 重复投票错误
	// 命名逻辑：Err + Vote + Repeated（重复投票错误）
	// 用于表示用户对同一帖子重复投相同的票
	ErrVoteRepeated = errors.New("不允许重复投票")
)

// CreatePost 创建帖子时初始化Redis数据结构
// 参数 postID: 帖子ID（int64类型）
// 参数 communityID: 社区ID（int64类型）
// 返回值: 错误信息，成功时返回nil
func CreatePost(postID, communityID int64) error {
	// pipeline: Redis事务流水线对象
	// 命名逻辑：pipeline（管道），表示批量执行Redis命令的管道
	pipeline := client.TxPipeline()

	// 帖子时间：将帖子ID和发布时间添加到时间排序集合
	pipeline.ZAdd(context.Background(), getRedisKey(KeyPostTimeZSet), &redis.Z{
		Score:  float64(time.Now().Unix()), // 发布时间戳作为分数
		Member: postID,                     // 帖子ID作为成员
	})

	// 帖子分数：将帖子ID和初始分数添加到分数排序集合
	pipeline.ZAdd(context.Background(), getRedisKey(KeyPostScoreZSet), &redis.Z{
		Score:  float64(time.Now().Unix()), // 初始分数等于发布时间戳
		Member: postID,                     // 帖子ID作为成员
	})

	// cKey: 社区键名
	// 命名逻辑：c + Key（community Key的缩写）
	// 生成格式：bluebell:community:{communityID}
	cKey := getRedisKey(KeyCommunitySetPF + strconv.Itoa(int(communityID)))

	// 将帖子ID添加到对应社区的集合中
	pipeline.SAdd(context.Background(), cKey, postID)

	// err: 错误变量
	// 命名逻辑：err（error的缩写），Go语言标准错误变量命名
	_, err := pipeline.Exec(context.Background())
	return err
}

// VoteForPost 处理用户对帖子的投票操作
// 参数 userID: 用户ID（字符串类型）
// 参数 postID: 帖子ID（字符串类型）
// 参数 value: 投票值（1=赞成，-1=反对，0=取消投票）
// 返回值: 错误信息，成功时返回nil
func VoteForPost(userID, postID string, value float64) error {
	// ==================== 第一步：判断投票时间限制 ====================
	// postTime: 帖子发布时间
	// 命名逻辑：post + Time（帖子时间）
	// 从Redis有序集合中获取帖子的发布时间戳
	postTime := client.ZScore(context.Background(), getRedisKey(KeyPostTimeZSet), postID).Val()

	// 检查帖子是否超过一周，超过则不允许投票
	if float64(time.Now().Unix())-postTime > oneWeekInSeconds {
		return ErrVoteTimeExpire
	}

	// ==================== 第二步：查询历史投票记录 ====================
	// ov: 原始投票值（Original Vote）
	// 命名逻辑：o + v（original vote的缩写）
	// 获取用户对该帖子的历史投票记录（1、-1、0或不存在）
	ov := client.ZScore(context.Background(), getRedisKey(KeyPostVotedZSetPF+postID), userID).Val()

	// 判断本次投票和历史投票是否一致，一致则返回重复投票错误
	if value == ov {
		return ErrVoteRepeated
	}

	// ==================== 第三步：计算分数变化 ====================
	// op: 操作方向（Operation）
	// 命名逻辑：o + p（operation的缩写）
	// 1表示分数增加，-1表示分数减少
	var op float64
	if value > ov {
		op = 1 // 分数增加（如：从-1改为1，或从0改为1）
	} else {
		op = -1 // 分数减少（如：从1改为-1，或从1改为0）
	}

	// diff: 投票差值（Difference）
	// 命名逻辑：diff（difference的缩写）
	// 计算两次投票的差值绝对值，用于计算分数变化量
	diff := math.Abs(ov - value)

	// ==================== 第四步：执行Redis事务 ====================
	// pipeline: Redis事务流水线对象
	// 命名逻辑：pipeline（管道），用于批量执行Redis命令
	pipeline := client.TxPipeline()

	// 更新帖子分数：根据投票差值计算分数变化
	// op*diff*scorePerVote: 分数变化量
	// op: 方向（+1或-1）
	// diff: 差值（1或2）
	// scorePerVote: 每票分数（432）
	pipeline.ZIncrBy(context.Background(), getRedisKey(KeyPostScoreZSet), op*diff*scorePerVote, postID)

	// ==================== 第五步：记录用户投票信息 ====================
	// 如果本次投票为0，表示取消投票，删除用户的投票记录
	if value == 0 {
		pipeline.ZRem(context.Background(), getRedisKey(KeyPostVotedZSetPF+postID), userID)
	} else {
		// 否则，添加或更新用户的投票记录
		// value: 投票值（1=赞成，-1=反对）
		// userID: 用户ID作为成员
		pipeline.ZAdd(context.Background(), getRedisKey(KeyPostVotedZSetPF+postID), &redis.Z{
			Score:  value,  // 投票值作为分数
			Member: userID, // 用户ID作为成员
		})
	}

	// ==================== 第六步：执行事务 ====================
	// err: 错误变量
	// 命名逻辑：err（error的缩写）
	_, err := pipeline.Exec(context.Background())
	return err
}
