package logic

import (
	"bluebell/dao/mysql"
	"bluebell/dao/redis"
	"bluebell/models"
	"bluebell/pkg/snowflake"

	"go.uber.org/zap"
)

func CreatePost(p *models.Post) (err error) {
	// 1. 生成post id
	p.ID = snowflake.GenID()
	// 2. 保存到数据库
	err = mysql.CreatePost(p)
	if err != nil {
		return err
	}
	err = redis.CreatePost(p.ID, p.CommunityID)
	return
	// 3. 返回
}

// GetPostById 根据帖子id查询帖子详情数据
func GetPostById(pid int64) (data *models.ApiPostDetail, err error) {
	// 查询并组合我们接口想用的数据
	post, err := mysql.GetPostById(pid)
	if err != nil {
		zap.L().Error("mysql.GetPostById(pid) failed",
			zap.Int64("pid", pid),
			zap.Error(err))
		return
	}
	// 根据作者id查询作者信息
	user, err := mysql.GetUserById(post.AuthorID)
	if err != nil {
		zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
			zap.Int64("author_id", post.AuthorID),
			zap.Error(err))
		return
	}
	// 根据社区id查询社区详细信息
	community, err := mysql.GetCommunityDetailByID(post.CommunityID)
	if err != nil {
		zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
			zap.Int64("community_id", post.CommunityID),
			zap.Error(err))
		return
	}
	// 接口数据拼接
	data = &models.ApiPostDetail{
		AuthorName:      user.Username,
		Post:            post,
		CommunityDetail: community,
	}
	return
}

// GetPostList 获取帖子列表
// 根据分页参数获取帖子列表，并关联查询作者信息和社区信息
// 参数 page: 页码，从1开始
// 参数 size: 每页大小，限制返回的帖子数量
// 返回值: 帖子详情列表和错误信息
func GetPostList(page, size int64) (data []*models.ApiPostDetail, err error) {
	// ==================== 第一步：从数据库获取基础帖子信息 ====================
	// 调用数据访问层获取分页的帖子列表
	// 这里只获取帖子的基本信息（标题、内容、作者ID等）
	posts, err := mysql.GetPostList(page, size)
	if err != nil {
		// 如果数据库查询失败，直接返回错误
		return nil, err
	}

	// ==================== 第二步：初始化结果切片 ====================
	// make([]*models.ApiPostDetail, 0, len(posts))
	// - 第一个参数：切片类型
	// - 第二个参数：初始长度（0）
	// - 第三个参数：容量（len(posts)）
	// 预分配容量可以提高性能，避免频繁的内存重新分配
	data = make([]*models.ApiPostDetail, 0, len(posts))

	// ==================== 第三步：遍历帖子列表，关联查询详细信息 ====================
	// range posts 遍历切片，i是索引，post是当前帖子对象
	for _, post := range posts {
		// ==================== 3.1：查询作者信息 ====================
		// 根据帖子中的作者ID查询用户的详细信息
		user, err := mysql.GetUserById(post.AuthorID)
		if err != nil {
			// 如果查询作者信息失败，记录错误日志但继续处理其他帖子
			// 这样即使某个作者信息查询失败，也不会影响整个列表的返回
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("author_id", post.AuthorID), // 记录失败的作者ID
				zap.Error(err))                        // 记录具体错误信息
			continue // 跳过当前帖子，继续处理下一个
		}

		// ==================== 3.2：查询社区信息 ====================
		// 根据帖子中的社区ID查询社区的详细信息
		community, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			// 如果查询社区信息失败，记录错误日志但继续处理
			// 注意：这里的错误日志信息有误，应该是GetCommunityDetailByID
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed", // 这里应该是GetCommunityDetailByID
				zap.Int64("community_id", post.CommunityID), // 记录失败的社区ID
				zap.Error(err)) // 记录具体错误信息
			continue // 跳过当前帖子，继续处理下一个
		}

		// ==================== 3.3：组装完整的帖子详情 ====================
		// 创建ApiPostDetail结构体，包含帖子的完整信息
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username, // 作者用户名
			Post:            post,          // 帖子基本信息（嵌入结构体）
			CommunityDetail: community,     // 社区详细信息（嵌入结构体）
		}

		// ==================== 3.4：添加到结果列表 ====================
		// append() 将帖子详情添加到结果切片中
		data = append(data, postDetail)
	}

	// ==================== 第四步：返回结果 ====================
	// 返回组装好的帖子详情列表
	return
}

func GetPostList2(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 2. 去redis查询id列表
	ids, err := redis.GetPostIDsInOrder(p)
	if err != nil {
		return
	}
	if len(ids) == 0 {
		zap.L().Warn("redis.GetPostIDsInOrder(p) return 0 data")
		return
	}
	zap.L().Debug("GetPostList2", zap.Any("ids", ids))
	// 3. 根据id去MySQL数据库查询帖子详细信息
	// 返回的数据还要按照我给定的id的顺序返回
	posts, err := mysql.GetPostListByIDs(ids)
	if err != nil {
		return
	}
	zap.L().Debug("GetPostList2", zap.Any("posts", posts))
	// 提前查询好每篇帖子的投票数
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		return
	}

	// 将帖子的作者及分区信息查询出来填充到帖子中
	for idx, post := range posts {
		// 根据作者id查询作者信息
		user, err := mysql.GetUserById(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue
		}
		// 根据社区id查询社区详细信息
		community, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("community_id", post.CommunityID),
				zap.Error(err))
			continue
		}
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)
	}
	return

}

func GetCommunityPostList(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 2. 去redis查询id列表
	ids, err := redis.GetCommunityPostIDsInOrder(p)
	if err != nil {
		return
	}
	if len(ids) == 0 {
		zap.L().Warn("redis.GetPostIDsInOrder(p) return 0 data")
		return
	}
	zap.L().Debug("GetCommunityPostIDsInOrder", zap.Any("ids", ids))
	// 3. 根据id去MySQL数据库查询帖子详细信息
	// 返回的数据还要按照我给定的id的顺序返回
	posts, err := mysql.GetPostListByIDs(ids)
	if err != nil {
		return
	}
	zap.L().Debug("GetPostList2", zap.Any("posts", posts))
	// 提前查询好每篇帖子的投票数
	voteData, err := redis.GetPostVoteData(ids)
	if err != nil {
		return
	}

	// 将帖子的作者及分区信息查询出来填充到帖子中
	for idx, post := range posts {
		// 根据作者id查询作者信息
		user, err := mysql.GetUserById(post.AuthorID)
		if err != nil {
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("author_id", post.AuthorID),
				zap.Error(err))
			continue
		}
		// 根据社区id查询社区详细信息
		community, err := mysql.GetCommunityDetailByID(post.CommunityID)
		if err != nil {
			zap.L().Error("mysql.GetUserById(post.AuthorID) failed",
				zap.Int64("community_id", post.CommunityID),
				zap.Error(err))
			continue
		}
		postDetail := &models.ApiPostDetail{
			AuthorName:      user.Username,
			VoteNum:         voteData[idx],
			Post:            post,
			CommunityDetail: community,
		}
		data = append(data, postDetail)
	}
	return
}

// GetPostListNew  将两个查询帖子列表逻辑合二为一的函数
func GetPostListNew(p *models.ParamPostList) (data []*models.ApiPostDetail, err error) {
	// 根据请求参数的不同，执行不同的逻辑。
	if p.CommunityID == 0 {
		// 查所有
		data, err = GetPostList2(p)
	} else {
		// 根据社区id查询
		data, err = GetCommunityPostList(p)
	}
	if err != nil {
		zap.L().Error("GetPostListNew failed", zap.Error(err))
		return nil, err
	}
	return
}
