package dateutils

import (
	"time"
)

var (
	YearOne      time.Time = time.Date(1900, time.January, 1, 0, 0, 0, 0, time.UTC)
	YearInfinity time.Time = time.Date(9999, 1, 1, 0, 0, 0, 1, time.UTC)
)
