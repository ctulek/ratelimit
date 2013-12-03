math.randomseed(os.time())

keys = {"test1", "atest1", "test2", "btest2",
  "test3", "ctest3", "test4", "dtest4", "test5", "etest5"}

methods = {"GET", "POST", "DELETE"}

function request()
   i = math.random(10)
   method = methods[math.random(3)]
   if method == "GET" or method == "DELETE" then
      return wrk.format(method, "/?key=" .. keys[i])
   else
      return wrk.format(method, "/?key=" .. keys[i] .. "&count=1&limit=10&duration=5s")
   end
end

