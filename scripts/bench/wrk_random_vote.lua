-- wrk_random_vote.lua
-- 随机点赞测试：模拟1万个帖子被随机用户点赞
-- 用法：wrk -t4 -c100 -d30s -s wrk_random_vote.lua http://127.0.0.1:8084

local tokens = {}
for line in io.lines("tokens.txt") do
    table.insert(tokens, line)
end

function request()
    local post_id = math.random(1, 10000)
    local token = tokens[math.random(#tokens)]
    local body = string.format('{"post_id":"%d","direction":1}', post_id)
    return wrk.format("POST", "/api/v1/vote", { ["Authorization"] = "Bearer " .. token, ["Content-Type"] = "application/json" }, body)
end 