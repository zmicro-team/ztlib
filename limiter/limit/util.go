package limit

import "time"

// RunValue run value int redis
type RunValue struct {
	Exist bool          //  false if key not exist.
	Count int64         // count value
	TTL   time.Duration // TTL in seconds
}
