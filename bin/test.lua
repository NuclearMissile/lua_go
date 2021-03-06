-- local function max(...) 
--     local args = {...}
--     local val, idx 
--     for i=1,#args do
--         if val == nil or args[i] > val then
--             val, idx = args[i], i
--         end
--     end
--     return val, idx
-- end

-- local function assert(v) 
--     if not v then fail() end
-- end

-- local v1 = max(3,9,7,128,35)
-- assert(v1 == 128)
-- local v2, i2 = max(3,9,7,128,35)
-- assert(v2 == 128 and i2 == 4)
-- local v3, i3 = max(max(3,9,7,128,35))
-- assert(v3==128 and i3 == 1)
-- local t = {max(3,9,7,128,35)}
-- assert(t[1]==128 and t[2]==4)


--print("hello, world")
--print(1)

--function newCounter()
--    local count = 0
--    return function()
--        count = count + 1
--        return count
--    end
--end
--
--c1 = newCounter()
--print(c1())
--print(c1())
--
--c2 = newCounter()
--print(c2())
--print(c1())
--print(c2())
--print(10)

--local mt = {}
--
--function vector(x, y)
--    local v = { x = x, y = y }
--    setmetatable(v, mt)
--    return v
--end
--
--mt.__add = function(v1, v2)
--    return vector(v1.x + v2.x, v1.y + v2.y)
--end
--
--mt.__sub = function(v1, v2)
--    return vector(v1.x - v2.x, v1.y - v2.y)
--end
--
--mt.__mul = function(v1, n)
--    return vector(v1.x * n, v1.y * n)
--end
--
--mt.__div = function(v1, n)
--    return vector(v1.x / n, v1.y / n)
--end
--mt.__len = function(v)
--    return (v.x * v.x + v.y * v.y) ^ 0.5
--end
--
--mt.__eq = function(v1, v2)
--    return v1.x == v2.x and v1.y == v2.y
--end
--
--mt.__index = function(v, k)
--    if k == 'print' then
--        print("[" .. v.x .. ", " .. v.y .. "]")
--    end
--end
--mt.__call = function(v)
--    print("[" .. v.x .. ", " .. v.y .. "]")
--end
--
--v1 = vector(1, 2)
--v1:print()
--v2 = vector(3, 4)
--v2:print()
--v3 = v1 * 2
--v3.print()
--v4 = v1 + v3
--v4.print()
--print(#v2)
--print(v1 == v2)
--print(v2 == vector(3, 4))
--v4()

--t = { b = 2, a = 1, c = 3 }
--for k, v in pairs(t) do
--    print(k, v)
--end
--print("")
--
--t.d = 4
--t.e = 5
--
--for k, v in pairs(t) do
--    print(k, v)
--end
--
--print("")
--
--for k, v in ipairs(t) do
--    print(k, v)
--end
--
--print("")
--
--t = { "a", "b", "c" }
--for k, v in ipairs(t) do
--    print(k, v)
--end

function div0(a, b)
    if b == 0 then
        error "div by 0"
    else
        return a / b
    end
end

function div1(a, b)
    return div0(a, b)
end

function div2(a, b)
    return div1(a, b)
end

aa = 5.0;
bb = 0x3f.5ap+9;
cc = 0xabc.defp-100;
print(bb)
print(cc)
dd = { a = 1, b = 2., aa = 3, 6};
print(dd[4])

ok, res = pcall(div2, 4, 2);
print(ok, res)
ok, res = pcall(div2, 5, 0);
print(ok, res)
ok, res = pcall(div2, {}, {});
print(ok, res)




