-- 投票接口压力测试脚本
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["Authorization"] = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2NjA0ODU4MTIxMzkzMzE1ODQsInVzZXJuYW1lIjoidXNlcm5hbWUiLCJleHAiOjE3ODI1Njk2OTgsImlzcyI6ImJsdWViZWxsIn0.iXgf03hUB67gLaMrkck_yviR4QPJ4hMfDTf6vJlclO4"

-- 随机选择帖子ID进行投票
function request()
    local post_ids = {"660491165283389440", "643115608531013632", "643115608312909824"}
    local directions = {"1", "-1"}
    
    local post_id = post_ids[math.random(#post_ids)]
    local direction = directions[math.random(#directions)]
    
    local body = string.format('{"post_id":"%s","direction":"%s"}', post_id, direction)
    return wrk.format(nil, "/api/v1/vote", nil, body)
end
