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
         return redis.call('DEL', KEYS[1])
     else
         return 0
   end
`

	//RenewalScript for renewal the lease
	RenewalScript =  `
	if redis.call('EXISTS', KEYS[1]) == 1
		then 
   			redis.call('PEXPIRE', KEYS[1], ARGV[1])
  			return 1 
	end
	return 0
`
)
