package redisrepo

// return 1  — успех
// return 0  — старого нет (not found)
// return -1 — новый уже существует (редкий случай коллизии) => считаем duplicate
const rotateTokenLua = `
local old = KEYS[1]
local newk = KEYS[2]
local val = ARGV[1]
local ttl_ms = tonumber(ARGV[2])

if redis.call('EXISTS', old) == 0 then
  return 0
end

local ok = redis.call('SET', newk, val, 'PX', ttl_ms, 'NX')
if ok then
  redis.call('DEL', old)
  return 1
else
  return -1
end
`
