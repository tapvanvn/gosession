# 
- A session_id is an int64, inscrease by 1 when issue a new session.
- Server has a table of svcode that has expirate time. Each svcode indicate by a svcode_index. A new svcode will be generated after a time duration.
- svcode_index = timestamp % 86400 / duration
- svcode has 2 part connected by ".". ```svcodeA.svcodeB```
- Server also has a table of step_salt (s). Each row correspond to a step_id and svcode_index. it's mean there will be ```(86400 / duration) * num_of_step``` number of step_salt(s) will be generated per day.
- step_id = session_id % num_of_step
- step_hash = ```md5(step_salt + loop(step_min + step_id, svcodeB) + step_salt)```

# 
- session_string_hash = ```sha256(svcodeA.step_hash)```
- sesison_string = ```svcode_index.session_id.session_string_hash```
- action_indicate is an integer. we can mount 1 by 1 to actually action or using a hash function on the request path to get these action_indicate (s).

#
- The mechanism has 2 part. The session provider and The session validator
- The session provider try to prevent the ddos attack by providing a limit quota to the ip that request issue the session_string.
- The session validator try to prevent the ddos attack by providing a limit quota to each action_indicate of each session_id.
- The both session provider and validator would not communicate with each others, they manage the data through on a common mempool. So they can run on separating services.
