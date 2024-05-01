package eventflow

type Subscription struct {
	ID    string
	event chan *Event
	done  chan struct{}
	topic *Topic
}

func (s *Subscription) Close() error {
	t := s.topic
	_ = t
	return s.topic.engine.CloseSubscription(s.topic.ID, s.ID)
}
