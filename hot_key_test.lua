-- 热key测试脚本：对同一个帖子进行高并发投票
wrk.method = "POST"
wrk.headers["Content-Type"] = "application/json"
wrk.headers["Authorization"] = "Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ1c2VyX2lkIjo2NjA0ODU4MTIxMzkzMzE1ODQsInVzZXJuYW1lIjoidXNlcm5hbWUiLCJleHAiOjE3ODI1Njk2OTgsImlzcyI6ImJsdWViZWxsIn0.iXgf03hUB67gLaMrkck_yviR4QPJ4hMfDTf6vJlclO4"

-- 固定对同一个帖子投票，模拟热key问题
function request()
    local post_id = "660491165283389440"  -- 固定帖子ID
    local directions = {"1", "-1"}
    local direction = directions[math.random(#directions)]
    
    local body = string.format('{"post_id":"%s","direction":"%s"}', post_id, direction)
    return wrk.format(nil, "/api/v1/vote", nil, body)
end 