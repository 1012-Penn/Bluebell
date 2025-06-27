// Package models 提供数据模型定义功能
// 定义系统中各种业务实体的数据结构，用于数据存储和传输
package models

import "time" // 导入时间包，用于时间类型定义

// 内存对齐概念说明：
// Go语言中结构体字段的排列顺序会影响内存占用
// 按照字段大小从大到小排列可以减少内存碎片，提高访问效率

// Post 帖子数据模型
// 定义帖子的基本信息结构，对应数据库中的帖子表
type Post struct {
	ID          int64     `json:"id,string" db:"post_id"`                            // 帖子ID，使用雪花算法生成，JSON序列化时转为字符串避免精度丢失
	AuthorID    int64     `json:"author_id" db:"author_id"`                          // 作者ID，关联用户表
	CommunityID int64     `json:"community_id" db:"community_id" binding:"required"` // 社区ID，关联社区表，必填字段
	Status      int32     `json:"status" db:"status"`                                // 帖子状态，如：1-正常，2-已删除
	Title       string    `json:"title" db:"title" binding:"required"`               // 帖子标题，必填字段
	Content     string    `json:"content" db:"content" binding:"required"`           // 帖子内容，必填字段
	CreateTime  time.Time `json:"create_time" db:"create_time"`                      // 帖子创建时间
}

// ApiPostDetail 帖子详情接口的响应结构体
// 用于API返回帖子详细信息，包含作者名、投票数等扩展信息
type ApiPostDetail struct {
	AuthorName       string             `json:"author_name"` // 作者名称，通过关联查询获取
	VoteNum          int64              `json:"vote_num"`    // 投票数量，包括赞成票和反对票的差值
	*Post                               // 嵌入帖子结构体，继承帖子的所有字段
	*CommunityDetail `json:"community"` // 嵌入社区信息，包含社区名称等详细信息
}

//区分Api模型和数据库模型,Api模型是给前端用的,数据库模型是给数据库用的
//数据库模型存储原始数据,Api模型存储处理后的数据
//数据库模型和Api模型是两个不同的结构体,但是它们之间有关系
