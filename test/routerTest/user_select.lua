
function event()
    local query = "SELECT id, user_name, password FROM users WHERE user_name = 'test3'"
    db_query(query)
end
