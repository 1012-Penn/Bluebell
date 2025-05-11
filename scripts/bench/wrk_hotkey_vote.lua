-- wrk_hotkey_vote.lua
-- 热点Key冲突测试：模拟3-5个热门帖子被大量用户点赞
-- 用法：wrk -t4 -c100 -d30s -s wrk_hotkey_vote.lua http://127.0.0.1:8084

local hot_posts = {1, 2, 3, 4, 5}
local tokens = {}
for line in io.lines("tokens.txt") do
    table.insert(tokens, line)
end

function request()
    local post_id = hot_posts[math.random(#hot_posts)]
    local token = tokens[math.random(#tokens)]
    local body = string.format('{"post_id":"%d","direction":1}', post_id)
    return wrk.format("POST", "/api/v1/vote", { ["Authorization"] = "Bearer " .. token, ["Content-Type"] = "application/json" }, body)
end 