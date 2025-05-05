-- single_request.lua

-- 请求头，包括 Authorization
local headers = {
    ["Authorization"] = "Bearer eyJhbGciOiJFUzI1NiIsInR5cCI6IkpXVCJ9.eyJVc2VySWQiOjIsImlzcyI6IkxvZ2luIiwiZXhwIjoxNzQ3MDQzMTc4LCJpYXQiOjE3NDY0MzgzNzh9.ApDg5oMJgmWi9knSts4cuhqce_8Vp5kqcW1wcA5zniJMe-4MeOjocpmS1wX96Q_a2vemy4UWM6I2bb6F_pvsqA"
}

-- 请求的 URL
local url = "http://127.0.0.1:8089/wgxdouyin/user/userInform/?query_user_id=3"

-- 返回请求对象
request = function()
    -- 使用 wrk.format 构建请求
    return wrk.format("GET", url, headers)
end

-- 响应处理函数
response = function(status, headers, body)
    -- 打印每个请求的响应状态码，响应时间和响应体（如果需要）
    print("Response Status: " .. status)
    print("Response Body: " .. body)
end
