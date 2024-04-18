package observer

/*
TODO: spawn new workers when tasks increase
*/

import (
	"context"
	"encoding/json"
	"sync"
	"time"

	"github.com/alexjoedt/echosight/dateutils"
	es "github.com/alexjoedt/echosight/internal"
	flow "github.com/alexjoedt/echosight/internal/eventflow"
	"github.com/alexjoedt/echosight/internal/logger"
	"github.com/alexjoedt/echosight/internal/notify"
	"github.com/google/uuid"
)

// executor is responsible to execute the checkers
type executor struct {
	id      string
	name    string
	sched   *Scheduler
	checker es.Checker

	mu       sync.Mutex
	lastRun  time.Time
	lastMail time.Time
	firstRun bool

	done     chan struct{}
	checkNow chan struct{}

	history *es.CheckHistory
}

type Scheduler struct {
	mu    sync.RWMutex
	tasks map[string]*executor

	taskPool chan *executor
	stop     chan struct{}

	detectorService  es.DetectorService
	metricService    es.MetricWriter
	eventHandler     *flow.Engine
	notifier         *notify.Notifier
	log              *logger.Logger
	schedulerRunning bool

	workerCount int
	workerWg    sync.WaitGroup
	schedulerWg sync.WaitGroup
}

func NewScheduler(ds es.DetectorService, ms es.MetricService, eh *flow.Engine, n *notify.Notifier) *Scheduler {
	workerCount := 3
	s := &Scheduler{
		mu:              sync.RWMutex{},
		tasks:           make(map[string]*executor),
		taskPool:        make(chan *executor, workerCount),
		workerCount:     workerCount,
		detectorService: ds,
		metricService:   ms,
		eventHandler:    eh,
		notifier:        n,
		stop:            make(chan struct{}, 1),
		log:             logger.New("Observer-Scheduler"),
	}

	return s
}

func (s *Scheduler) AddDetector(detectorID uuid.UUID) (*executor, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	d, err := s.detectorService.GetByID(ctx, detectorID)
	if err != nil {
		return nil, err
	}

	checker, err := d.GetChecker()
	if err != nil {
		return nil, err
	}

	task := &executor{
		id:       d.ID.String(),
		name:     d.Name,
		checker:  checker,
		sched:    s,
		done:     make(chan struct{}, 1),
		checkNow: make(chan struct{}, 1),
		lastRun:  dateutils.YearOne,
		lastMail: dateutils.YearOne,
		firstRun: true,
		history:  &es.CheckHistory{Results: make([]*es.Result, 3)},
	}

	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.tasks[d.ID.String()]; !ok {
		s.tasks[d.ID.String()] = task
	}

	// create an topic for each detector
	_, err = s.eventHandler.NewTopic(d.ID.String())
	if err != nil {
		return nil, err
	}

	d.Active = true
	err = s.detectorService.Update(ctx, d)
	if err != nil {
		s.RemoveDetector(d.ID)
		return nil, err
	}

	return task, nil
}

func (s *Scheduler) RemoveDetector(detectorID uuid.UUID) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	d, err := s.detectorService.GetByID(ctx, detectorID)
	if err != nil {
		return err
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	_, ok := s.tasks[detectorID.String()]
	if !ok {
		s.log.Errorf("no detector job is running with provided id")
	} else {
		delete(s.tasks, detectorID.String())
	}

	// even if no detector is registered, update the state
	d.Active = false
	d.State = es.StateInactive
	err = s.detectorService.Update(ctx, d)
	if err != nil {
		return err
	}

	err = s.eventHandler.CloseTopic(detectorID.String())
	if err != nil {
		s.log.Errorf("failed to close detector topic: %v", err)
	}

	return nil
}

func (s *Scheduler) AddDetectors(ds ...*es.Detector) error {
	for _, d := range ds {
		_, err := s.AddDetector(d.ID)
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *Scheduler) Start() {
	if s.IsRunning() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.schedulerWg.Add(1)
	s.schedulerRunning = true

	s.taskPool = make(chan *executor, s.workerCount)
	for i := 0; i < s.workerCount; i++ {
		s.workerWg.Add(1)
		go s.worker(i)
	}

	go func() {
		ticker := time.NewTicker(time.Second * 1)
		defer func() {
			s.schedulerWg.Done()
			ticker.Stop()

			s.mu.Lock()
			s.schedulerRunning = false
			s.mu.Unlock()
		}()

		for {
			select {
			case <-ticker.C:
				now := time.Now()
				s.mu.RLock()
				for _, checkerTasks := range s.tasks {
					checkerTasks.mu.Lock()
					if now.Sub(checkerTasks.lastRun) >= checkerTasks.checker.Interval() {
						// lastRun must set immediatley here,
						// because the execution could be take more then the scheduler Tick (1s)
						// and then the scheduler would start every second a new execution as long the first run is not finished
						checkerTasks.lastRun = time.Now()

						s.taskPool <- checkerTasks
					}
					checkerTasks.mu.Unlock()
				}
				s.mu.RUnlock()
			case <-s.stop:
				ticker.Stop()
				close(s.taskPool)
				return
			}
		}
	}()

}

// Stop, blocks until scheduler and worker done
func (s *Scheduler) Stop() {
	if !s.IsRunning() {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()

	s.stop <- struct{}{}
	s.schedulerWg.Wait()
	s.workerWg.Wait()
	s.schedulerRunning = false
}

func (s *Scheduler) IsRunning() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.schedulerRunning
}

func (s *Scheduler) worker(i int) {
	defer s.workerWg.Done()
	s.log.Infof("Start worker '%d'", i)
	for t := range s.taskPool {
		t.runCheck()
		t.firstRun = false
	}
	s.log.Infof("worker '%d' done", i)
}

func (t *executor) CheckNow() error {
	t.mu.Lock()
	defer t.mu.Unlock()
	t.runCheck()
	return nil
}

func (t *executor) runCheck() {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*10)
	defer cancel()

	result := t.checker.Check(ctx)
	if result.Error() != nil {
		t.sched.log.Errorf("check failed: %v", result.Error())
	}

	t.processResult(result, t.checker.Detector())
}

// processResul
func (t *executor) processResult(result *es.Result, detector *es.Detector) error {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	result.Host = detector.HostName
	result.Detector = detector.Name

	t.history.AddResult(result)

	detector.LastCheckedAt = time.Now()
	detector.State = result.State
	err := t.sched.detectorService.Update(ctx, detector)
	if err != nil {
		t.sched.log.Errorf("failed to update detector after check: %v", err)
	}

	if result.Metric != nil {
		err = t.sched.metricService.Write(ctx, result.Metric)
		if err != nil {
			t.sched.log.Errorf("failed to write metrics to influx: %v", err)
		}
	}

	if t.shouldNotify(result) {
		t.lastMail = time.Now()
		err = t.sched.notifier.Send(ctx, result)
		if err != nil {
			t.sched.log.Errorf("failed to send notifications: %v", err)
		}
	}

	eventPayload := es.ResultEvent{
		HostID:       detector.HostID.String(),
		HostName:     detector.Name,
		DetectorID:   detector.ID.String(),
		DetectorName: detector.Name,
		CheckResult:  result,
	}

	payload, err := json.Marshal(eventPayload)
	if err != nil {
		t.sched.log.Errorf("failed to marshal payload: %v", err)
	} else {
		// INFO: publish blocks until the event is read from all subscribers
		t.sched.eventHandler.Publish(detector.ID.String(), &flow.Event{
			Type:    es.EventCheckResult,
			Payload: payload,
		})
	}

	return nil
}

func (e *executor) shouldNotify(r *es.Result) bool {
	if e.firstRun && r.State != es.StateOK {
		return true
	}

	if !e.firstRun && e.history.StateChanged() {
		return true
	}

	return false
}
