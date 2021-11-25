package clock

import "time"

type Clock interface {
	Now() time.Time
}

type Frozen struct {
	now time.Time
}

func Freeze(now time.Time) Frozen {
	return Frozen{now: now}
}

func (f Frozen) Now() time.Time { return f.now }
