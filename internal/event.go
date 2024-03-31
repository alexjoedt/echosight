package echosight

const (
	EventCheckResult = "check_result"
)

type ResultEvent struct {
	HostID       string
	DetectorID   string
	HostName     string
	DetectorName string
	CheckResult  *Result
}
