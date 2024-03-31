package echosight

type DetectorType string

func (s DetectorType) String() string {
	return string(s)
}

var (
	DetectorHTTP     DetectorType = "http"
	DetectorPostgres DetectorType = "psql"
	DetectorAgent    DetectorType = "agent"
)
