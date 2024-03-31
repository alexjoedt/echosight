package http

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/alexjoedt/echosight/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) handlerCreateDetector(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	host, err := s.checkForHost(ctx, r)
	if err != nil {
		return err
	}

	var input struct {
		Name     string                  `json:"name"`
		Type     echosight.DetectorType  `json:"type"`
		Timeout  echosight.Duration      `json:"timeout"`
		Interval echosight.Duration      `json:"interval"`
		Tags     []string                `json:"tags"`
		Config   echosight.CheckerConfig `json:"config"`
	}

	err = readJSON(r, &input)
	if err != nil {
		s.log.Errorc("failed to read detector payload", err)
		return err
	}

	detector := echosight.Detector{
		Name:     input.Name,
		HostID:   host.ID,
		HostName: host.Name,
		Active:   false,
		State:    "inactive",
		Type:     input.Type,
		Tags:     input.Tags,
		Timeout:  input.Timeout,
		Interval: input.Interval,
		Config:   input.Config,
	}

	v := validator.New()
	echosight.ValidateDetector(v, &detector)
	if !v.Valid() {
		return echosight.ErrInvalidf("invalid detector payload").WithData(v.Errors)
	}

	err = s.DetectorService.Create(ctx, &detector)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status:  StatusOK,
		Message: "detector created",
		Data: W{
			"detector": detector,
		},
	})
}

func (s *Server) handlerGetDetectorByID(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	id := chi.URLParam(r, "detectorID")
	if id == "" {
		return echosight.ErrInvalidf("no detector ID provided")
	}

	detectorID, err := uuid.Parse(id)
	if err != nil {
		s.log.Errorc("invalid id: %v", err)
		return echosight.ErrInvalidf("invalid ID '%s'", id)
	}

	detector, err := s.DetectorService.GetByID(ctx, detectorID)
	if err != nil {
		return err
	}

	// always return only 10 metric points
	timeRange := 60 * ((10 * time.Duration(detector.Interval).Seconds()) / 60)
	str := fmt.Sprintf("-%ds", int(timeRange))
	metrics, err := s.MetricReader.Read(ctx, detector.MetricFiler(str))
	if err != nil {
		s.log.Errorf("failed to read metrics")
	} else {
		detector.Metrics = metrics
	}

	return writeJSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data: W{
			"detector": detector,
		},
	})
}

func (s *Server) handlerGetDetectors(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	host, err := s.checkForHost(ctx, r)
	if err != nil {
		return err
	}

	// TODO: read filter params
	detectorFilter := filter.NewDefaultDetectorFilter()
	detectorFilter.HostID = &host.ID

	detectors, err := s.DetectorService.List(ctx, detectorFilter)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data: W{
			"detectors":  detectors,
			"pagination": detectorFilter.Pagination,
		},
	})
}

func (s *Server) handlerUpdateDetector(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	id := chi.URLParam(r, "detectorID")
	if id == "" {
		return echosight.ErrInvalidf("no detector ID provided")
	}

	detectorID, err := uuid.Parse(id)
	if err != nil {
		s.log.Errorc("invalid id: %v", err)
		return echosight.ErrInvalidf("invalid ID '%s'", id)
	}

	err = s.Scheduler.RemoveDetector(detectorID)
	if err != nil {
		return err
	}

	// The detector will be activated and deactivated through a seperate route.
	var input struct {
		Name     *string                  `json:"name"`
		Type     *string                  `json:"type"`
		Timeout  *echosight.Duration      `json:"timeout"`
		Interval *echosight.Duration      `json:"interval"`
		Tags     []string                 `json:"tags"`
		Config   *echosight.CheckerConfig `json:"config"`
	}

	err = readJSON(r, &input)
	if err != nil {
		s.log.Errorc("failed to read detector payload", err)
		return err
	}

	detector, err := s.DetectorService.GetByID(ctx, detectorID)
	if err != nil {
		return err
	}

	if input.Name != nil {
		detector.Name = *input.Name
	}

	if input.Type != nil {
		detector.Type = echosight.DetectorType(*input.Type)
	}

	if input.Timeout != nil {
		detector.Timeout = *input.Timeout
	}

	if input.Interval != nil {
		detector.Interval = *input.Interval
	}

	if input.Tags != nil {
		detector.Tags = input.Tags
	}

	if input.Config != nil {
		detector.Config = *input.Config
	}

	detector.UpdatedAt = time.Now()

	v := validator.New()
	echosight.ValidateDetector(v, detector)
	if !v.Valid() {
		return echosight.ErrInvalidf("invalid detector payload").WithData(v.Errors)
	}

	err = s.DetectorService.Update(ctx, detector)
	if err != nil {
		return err
	}

	_, err = s.Scheduler.AddDetector(detectorID)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status:  StatusOK,
		Message: "detector updated",
		Data: W{
			"detector": detector,
		},
	})
}

func (s *Server) handlerDeleteDetectorByID(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	id := chi.URLParam(r, "detectorID")
	if id == "" {
		return echosight.ErrInvalidf("no detector ID provided")
	}

	detectorID, err := uuid.Parse(id)
	if err != nil {
		s.log.Errorc("invalid id: %v", err)
		return echosight.ErrInvalidf("invalid ID '%s'", id)
	}

	detector, err := s.DetectorService.DeleteByID(ctx, detectorID)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data: W{
			"detector": detector,
		},
	})
}

func (s *Server) handlerActivateDetector(w http.ResponseWriter, r *http.Request) error {

	id, err := ReadUUIDParam(r, "detectorID")
	if err != nil {
		return err
	}

	_, err = s.Scheduler.AddDetector(id)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status:  StatusOK,
		Message: "detector activated",
	})
}

func (s *Server) handlerDeactivateDetector(w http.ResponseWriter, r *http.Request) error {
	id, err := ReadUUIDParam(r, "detectorID")
	if err != nil {
		return err
	}

	err = s.Scheduler.RemoveDetector(id)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status:  StatusOK,
		Message: "detector deactivated",
	})
}

func (s *Server) handlerObserverStatus(w http.ResponseWriter, r *http.Request) error {

	var input struct {
		Command string `json:"command"`
	}

	err := readJSON(r, &input)
	if err != nil {
		return err
	}

	input.Command = strings.ToLower(input.Command)

	var msg string

	if input.Command == "start" {
		if !s.Scheduler.IsRunning() {
			s.Scheduler.Start()
			msg = "monitoring started"
		} else {
			msg = "monitoring is already running"
		}
	} else if input.Command == "stop" {
		if s.Scheduler.IsRunning() {
			s.Scheduler.Stop()
			msg = "monitoring stopped"
		} else {
			msg = "monitoring is already stopped"
		}
	} else {
		msg = "invalid command, use 'start' or 'stop'"
	}

	return writeJSON(w, http.StatusOK, Response{
		Status:  StatusOK,
		Message: msg,
	})
}

func (s *Server) checkForHost(ctx context.Context, r *http.Request) (*echosight.Host, error) {
	hostID, err := ReadUUIDParam(r, "hostID")
	if err != nil {
		return nil, err
	}

	// check if a host exists with the given ID
	h, err := s.HostService.GetByID(ctx, hostID)
	if err != nil {
		return nil, echosight.ErrNotfoundf("no host found with ID: '%s'", hostID.String())
	}

	return h, nil
}
