package echosight

type CheckHistory struct {
	Results []*Result
}

func (ch *CheckHistory) AddResult(result *Result) {
	ch.Results = append(ch.Results, result)
	if len(ch.Results) > 3 {
		ch.Results = ch.Results[len(ch.Results)-3:]
	}
}

func (ch *CheckHistory) StateChanged() bool {
	last := len(ch.Results) - 1
	if last <= 0 {
		return false
	}

	if ch.Results[last-1] == nil {
		return false
	}

	if ch.Results[last] == nil {
		return false
	}

	return ch.Results[last-1].State != ch.Results[last].State
}

// WarnOrCritical evaluates if all results in history
// are warn or critical
func (ch *CheckHistory) WarnOrCritical() bool {
	for _, h := range ch.Results {
		if h.State == StateOK {
			return false
		}
	}

	return true
}
