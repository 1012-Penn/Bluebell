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
func GetPostList(page, size int64) (posts []*models.Post, err error) {
	sqlStr := `select 
	post_id, title, content, author_id, community_id, create_time
	from post
	ORDER BY create_time
	DESC
	limit ?,?
	`
	posts = make([]*models.Post, 0, 2) // 不要写成make([]*models.Post, 2)
	err = db.Select(&posts, sqlStr, (page-1)*size, size)
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

// InsertPostVote 插入点赞记录
func InsertPostVote(userID int64, postID string, voteType int8) error {
	sqlStr := `INSERT INTO post_vote (user_id, post_id, vote_type) VALUES (?, ?, ?)
	ON DUPLICATE KEY UPDATE vote_type = VALUES(vote_type), create_time = CURRENT_TIMESTAMP`
	_, err := db.Exec(sqlStr, userID, postID, voteType)
	return err
}

// DeletePostVote 删除点赞记录（如果需要取消点赞时删除记录，可选）
func DeletePostVote(userID int64, postID string) error {
	sqlStr := `DELETE FROM post_vote WHERE user_id = ? AND post_id = ?`
	_, err := db.Exec(sqlStr, userID, postID)
	return err
}
