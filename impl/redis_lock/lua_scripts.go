package redis_lock

var (
	//CheckExpireScript for checking the key is exits or not
	CheckExpireScript = `
   local val = redis.call('GET', KEYS[1])
   if val == ARGV[1]
     then 
         return 1
     else
         return 0
   end
`

	//UnlockScript for unlocking redis lock
	UnlockScript = `
   if redis.call('GET', KEYS[1]) == ARGV[1]
     then 
		 redis.call('PUBLISH', KEYS[2], ARGV[2])  -- awake waiting client
         redis.call('DEL', KEYS[1]) 
		 return 0
     else
		 redis.call('PUBLISH', KEYS[2], ARGV[2])  -- awake waiting client
         return 1
   end
`

	//RenewalScript for renewal the lease
	RenewalScript = `
		local val =  redis.call('GET', KEYS[1])
		if val == false  -- false means key is not exist
		then
			redis.call('PEXPIRE', KEYS[1], ARGV[2])
  			return 0
		else
			if val == ARGV[1] 
			then
				redis.call('PEXPIRE', KEYS[1], ARGV[2])
			    return 1
			else 
				return 2
			end
		end
`
)
