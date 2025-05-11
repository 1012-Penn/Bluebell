-- wrk_mixed_vote.lua
-- 双模式混合测试：50%概率点赞热门帖子，50%概率随机点赞
-- 用法：wrk -t4 -c100 -d30s -s wrk_mixed_vote.lua http://127.0.0.1:8084

local hot_posts = {1, 2, 3, 4, 5}
local tokens = {}
for line in io.lines("tokens.txt") do
    table.insert(tokens, line)
end

function request()
    local post_id
    if math.random() < 0.5 then
        post_id = hot_posts[math.random(#hot_posts)]
    else
        post_id = math.random(1, 10000)
    end
    local token = tokens[math.random(#tokens)]
    local body = string.format('{"post_id":"%d","direction":1}', post_id)
    return wrk.format("POST", "/api/v1/vote", { ["Authorization"] = "Bearer " .. token, ["Content-Type"] = "application/json" }, body)
end 