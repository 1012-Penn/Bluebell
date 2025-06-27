package mysql

import (
	"bluebell/models"
	"strings"

	"github.com/jmoiron/sqlx"
)

// CreatePost 创建帖子
func CreatePost(p *models.Post) (err error) {
	sqlStr := `insert into post(
	post_id, title, content, author_id, community_id)
	values (?, ?, ?, ?, ?)
	`
	_, err = db.Exec(sqlStr, p.ID, p.Title, p.Content, p.AuthorID, p.CommunityID)
	return
}

// GetPostById 根据id查询单个贴子数据
func GetPostById(pid int64) (post *models.Post, err error) {
	post = new(models.Post)
	sqlStr := `select
	post_id, title, content, author_id, community_id, create_time
	from post
	where post_id = ?
	`
	err = db.Get(post, sqlStr, pid)
	return
}

// GetPostList 查询帖子列表函数
// 根据分页参数从数据库获取帖子列表，按创建时间倒序排列
// 参数 page: 页码，从1开始
// 参数 size: 每页大小，限制返回的帖子数量
// 返回值: 帖子列表和错误信息
func GetPostList(page, size int64) (posts []*models.Post, err error) {
	// ==================== 第一步：构建SQL查询语句 ====================
	// 使用反引号定义多行SQL字符串，保持格式清晰
	sqlStr := `select 
	post_id, title, content, author_id, community_id, create_time
	from post
	ORDER BY create_time
	DESC
	limit ?,?
	`
	// SQL语句说明：
	// - select: 选择指定字段
	// - from post: 从post表查询
	// - ORDER BY create_time DESC: 按创建时间倒序排列（最新的在前）
	// - limit ?,?: 分页限制，第一个?是偏移量，第二个?是限制数量

	// ==================== 第二步：初始化结果切片 ====================
	// make([]*models.Post, 0, 2) 创建一个切片
	// - 第一个参数：切片类型 []*models.Post
	// - 第二个参数：初始长度 0（空切片）
	// - 第三个参数：容量 2（预分配2个元素的空间）
	// 注意：这里容量设置为2可能不够，应该根据实际size参数设置
	posts = make([]*models.Post, 0, 2) // 不要写成make([]*models.Post, 2)
	// 注释说明：如果写成make([]*models.Post, 2)，会创建一个长度为2的切片，
	// 包含2个nil指针，这不是我们想要的

	// ==================== 第三步：执行数据库查询 ====================
	// db.Select() 执行查询并将结果映射到posts切片
	// 参数说明：
	// - &posts: 结果映射的目标切片（指针）
	// - sqlStr: SQL查询语句
	// - (page-1)*size: 偏移量，计算要跳过的记录数
	// - size: 限制返回的记录数
	//
	// 分页计算示例：
	// page=1, size=10: 偏移量=(1-1)*10=0，返回前10条
	// page=2, size=10: 偏移量=(2-1)*10=10，返回第11-20条
	// page=3, size=10: 偏移量=(3-1)*10=20，返回第21-30条
	err = db.Select(&posts, sqlStr, (page-1)*size, size)

	// ==================== 第四步：返回结果 ====================
	return
}

// GetPostListByIDs 根据给定的id列表查询帖子数据
func GetPostListByIDs(ids []string) (postList []*models.Post, err error) {
	sqlStr := `select post_id, title, content, author_id, community_id, create_time
	from post
	where post_id in (?)
	order by FIND_IN_SET(post_id, ?)
	`
	// https: //www.liwenzhou.com/posts/Go/sqlx/
	query, args, err := sqlx.In(sqlStr, ids, strings.Join(ids, ","))
	if err != nil {
		return nil, err
	}
	query = db.Rebind(query)
	err = db.Select(&postList, query, args...) // !!!!!!
	return
}
