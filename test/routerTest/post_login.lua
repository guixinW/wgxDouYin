wrk.method = "POST"
wrk.headers["Content-Type"] = "application/x-www-form-urlencoded"
users = {
    { username = "test1", password = "12345" },
    { username = "test2", password = "12345" },
    { username = "test3", password = "12345" },
    { username = "test4", password = "12345" },
    { username = "test5", password = "12345" },
}

function request()
    local user = users[math.random(#users)]
    local body = "username=" .. user.username .. "&password=" .. user.password
    return wrk.format(nil, nil, nil, body)
end
