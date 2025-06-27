-- ==================== 数据库表结构设计说明 ====================
-- 本文件包含bluebell项目的所有数据库表结构
-- 设计原则：高性能、易扩展、数据一致性
-- 索引策略：主键索引 + 业务索引 + 唯一索引

-- ==================== 用户表 (user) ====================
-- 设计思路：
-- 1. 使用bigint类型存储用户ID，支持大量用户
-- 2. 用户名和用户ID都设置唯一索引，防止重复
-- 3. 密码字段存储加密后的密码
-- 4. 性别使用tinyint，节省存储空间
-- 5. 自动记录创建和更新时间
CREATE TABLE `user` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,           -- 自增主键，用于内部关联
    `user_id` bigint(20) NOT NULL,                     -- 用户ID，业务主键，全局唯一
    `username` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,  -- 用户名，用于登录
    `password` varchar(64) COLLATE utf8mb4_general_ci NOT NULL,  -- 加密后的密码
    `email` varchar(64) COLLATE utf8mb4_general_ci,              -- 邮箱，可选字段
    `gender` tinyint(4) NOT NULL DEFAULT '0',          -- 性别：0=未知，1=男，2=女
    `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP,      -- 创建时间，自动设置
    `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,  -- 更新时间，自动更新
    PRIMARY KEY (`id`),                                 -- 主键索引，用于内部关联
    UNIQUE KEY `idx_username` (`username`) USING BTREE, -- 用户名唯一索引，防止重复注册
    UNIQUE KEY `idx_user_id` (`user_id`) USING BTREE   -- 用户ID唯一索引，业务主键
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ==================== 社区表 (community) ====================
-- 设计思路：
-- 1. 社区ID使用int类型，支持最多21亿个社区
-- 2. 社区名称唯一，防止重复创建
-- 3. 简介字段用于描述社区特点
-- 4. 自动记录创建和更新时间
DROP TABLE IF EXISTS `community`;
CREATE TABLE `community` (
     `id` int(11) NOT NULL AUTO_INCREMENT,             -- 自增主键，用于内部关联
     `community_id` int(10) unsigned NOT NULL,         -- 社区ID，业务主键
     `community_name` varchar(128) COLLATE utf8mb4_general_ci NOT NULL,  -- 社区名称
     `introduction` varchar(256) COLLATE utf8mb4_general_ci NOT NULL,    -- 社区简介
     `create_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP,         -- 创建时间
     `update_time` timestamp NOT NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP,  -- 更新时间
     PRIMARY KEY (`id`),                               -- 主键索引
     UNIQUE KEY `idx_community_id` (`community_id`),   -- 社区ID唯一索引
     UNIQUE KEY `idx_community_name` (`community_name`) -- 社区名称唯一索引，防止重复
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ==================== 社区初始数据 ====================
-- 预置4个热门社区，覆盖不同技术领域
INSERT INTO `community` VALUES ('1', '1', 'Go', 'Golang', '2016-11-01 08:10:10', '2016-11-01 08:10:10');
INSERT INTO `community` VALUES ('2', '2', 'leetcode', '刷题刷题刷题', '2020-01-01 08:00:00', '2020-01-01 08:00:00');
INSERT INTO `community` VALUES ('3', '3', 'CS:GO', 'Rush B。。。', '2018-08-07 08:30:00', '2018-08-07 08:30:00');
INSERT INTO `community` VALUES ('4', '4', 'LOL', '欢迎来到英雄联盟!', '2016-01-01 08:00:00', '2016-01-01 08:00:00');

-- ==================== 帖子表 (post) ====================
-- 设计思路：
-- 1. 使用bigint存储帖子ID，支持海量帖子
-- 2. 标题和内容使用utf8mb4字符集，支持emoji
-- 3. 作者ID关联用户表
-- 4. 社区ID关联社区表
-- 5. 状态字段支持帖子管理（如删除、置顶等）
-- 6. 创建复合索引优化查询性能
DROP TABLE IF EXISTS `post`;
CREATE TABLE `post` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,           -- 自增主键，用于内部关联
    `post_id` bigint(20) NOT NULL COMMENT '帖子id',    -- 帖子ID，业务主键，全局唯一
    `title` varchar(128) COLLATE utf8mb4_general_ci NOT NULL COMMENT '标题',  -- 帖子标题
    `content` varchar(8192) COLLATE utf8mb4_general_ci NOT NULL COMMENT '内容',  -- 帖子内容，支持长文本
    `author_id` bigint(20) NOT NULL COMMENT '作者的用户id',  -- 作者ID，关联用户表
    `community_id` bigint(20) NOT NULL COMMENT '所属社区',   -- 社区ID，关联社区表
    `status` tinyint(4) NOT NULL DEFAULT '1' COMMENT '帖子状态',  -- 状态：1=正常，0=删除，2=置顶等
    `create_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP COMMENT '创建时间',  -- 发布时间
    `update_time` timestamp NULL DEFAULT CURRENT_TIMESTAMP ON UPDATE CURRENT_TIMESTAMP COMMENT '更新时间',  -- 最后修改时间
    PRIMARY KEY (`id`),                               -- 主键索引
    UNIQUE KEY `idx_post_id` (`post_id`),             -- 帖子ID唯一索引
    KEY `idx_author_id` (`author_id`),                -- 作者ID索引，优化按作者查询
    KEY `idx_community_id` (`community_id`)           -- 社区ID索引，优化按社区查询
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;

-- ==================== 投票表 (post_vote) ====================
-- 设计思路：
-- 1. 记录用户对帖子的投票历史
-- 2. 使用复合唯一索引防止重复投票
-- 3. vote_type字段：1=赞成，-1=反对
-- 4. 支持投票历史查询和统计
-- 5. 与Redis投票数据保持同步
DROP TABLE IF EXISTS `post_vote`;
CREATE TABLE `post_vote` (
    `id` bigint(20) NOT NULL AUTO_INCREMENT,           -- 自增主键
    `user_id` bigint(20) NOT NULL COMMENT '用户ID',    -- 投票用户ID
    `post_id` bigint(20) NOT NULL COMMENT '帖子ID',    -- 被投票的帖子ID
    `vote_type` tinyint(4) YES DEFAULT 1 COMMENT '投票类型',  -- 投票类型：1=赞成，-1=反对
    `create_time` timestamp YES DEFAULT CURRENT_TIMESTAMP DEFAULT_GENERATED COMMENT '创建时间',  -- 投票时间
    PRIMARY KEY (`id`),                               -- 主键索引
    UNIQUE KEY `uk_post_user` (`post_id`, `user_id`) COMMENT '防止重复投票'  -- 复合唯一索引，确保每个用户对每个帖子只能投一票
) ENGINE=InnoDB DEFAULT CHARSET=utf8mb4 COLLATE=utf8mb4_general_ci;