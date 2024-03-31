package http

import (
	"net/http"
	"time"

	echosight "github.com/alexjoedt/echosight/internal"
	"github.com/alexjoedt/echosight/internal/filter"
	"github.com/alexjoedt/echosight/internal/validator"
	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (s *Server) handlerCreateHost(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	var input struct {
		Name        string                `json:"name"`
		AddressType echosight.AddressType `json:"addressType"`
		Address     string                `json:"address"`
		Agent       bool                  `json:"agent"`
		Location    string                `json:"location"`
		OS          string                `json:"os"`
		Tags        []string              `json:"tags"`
	}

	err := readJSON(r, &input)
	if err != nil {
		s.log.Errorc("failed to read host payload", err)
		return err
	}

	host := echosight.Host{
		Name:        input.Name,
		Address:     input.Address,
		AddressType: input.AddressType,
		Location:    input.Location,
		OS:          input.OS,
		Agent:       input.Agent,
		Active:      false,
		State:       "inactive",
		Tags:        input.Tags,
	}

	v := validator.New()
	echosight.ValidateHost(v, &host)
	if !v.Valid() {
		return echosight.ErrInvalidf("invalid host payload").WithData(v.Errors)
	}

	// TODO: if agent true, crate or activate AgentDetectors

	err = s.HostService.Create(ctx, &host)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status:  StatusOK,
		Message: "host created",
		Data: W{
			"host": host,
		},
	})
}

func (s *Server) handlerGetHostByID(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	id := chi.URLParam(r, "hostID")
	if id == "" {
		return echosight.ErrInvalidf("no host ID provided")
	}

	hostID, err := uuid.Parse(id)
	if err != nil {
		s.log.Errorc("invalid id: %v", err)
		return echosight.ErrInvalidf("invalid ID '%s'", id)
	}

	host, err := s.HostService.GetByID(ctx, hostID)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data: W{
			"host": host,
		},
	})
}

func (s *Server) handlerGetHosts(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	// TODO: read filter params
	hostFilter := filter.NewDefaultHostFilter()

	hosts, err := s.HostService.List(ctx, hostFilter)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data: W{
			"hosts":      hosts,
			"pagination": hostFilter.Pagination,
		},
	})
}

func (s *Server) handlerUpdateHost(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	id := chi.URLParam(r, "hostID")
	if id == "" {
		return echosight.ErrInvalidf("no host ID provided")
	}

	hostID, err := uuid.Parse(id)
	if err != nil {
		s.log.Errorc("invalid id: %v", err)
		return echosight.ErrInvalidf("invalid ID '%s'", id)
	}

	var input struct {
		Name        *string  `json:"name"`
		Address     *string  `json:"address"`
		IPv4        *string  `json:"addressType"`
		AddressType *string  `json:"ipv6"`
		Agent       *bool    `json:"agent"`
		Location    *string  `json:"location"`
		Active      *bool    `json:"active"`
		OS          *string  `json:"os"`
		Tags        []string `json:"tags"`
	}

	err = readJSON(r, &input)
	if err != nil {
		s.log.Errorc("failed to read host payload", err)
		return err
	}

	host, err := s.HostService.GetByID(ctx, hostID)
	if err != nil {
		return err
	}

	if input.Name != nil {
		host.Name = *input.Name
	}

	if input.Address != nil {
		host.Address = *input.Address
	}

	if input.AddressType != nil {
		host.AddressType = echosight.AddressType((*input.AddressType))
	}

	if input.Location != nil {
		host.Location = *input.Location
	}

	if input.OS != nil {
		host.OS = *input.OS
	}

	if input.Active != nil {
		host.Active = *input.Active
	}

	if input.Agent != nil {
		host.Agent = *input.Agent
	}

	if input.Tags != nil {
		host.Tags = input.Tags
	}

	host.UpdatedAt = time.Now()

	v := validator.New()
	echosight.ValidateHost(v, host)
	if !v.Valid() {
		return echosight.ErrInvalidf("invalid host payload").WithData(v.Errors)
	}

	err = s.HostService.Update(ctx, host)
	if err != nil {
		return err
	}

	// TODO: if agent true, crate or activate AgentDetectors

	return writeJSON(w, http.StatusOK, Response{
		Status:  StatusOK,
		Message: "host updated",
		Data: W{
			"host": host,
		},
	})
}

func (s *Server) handlerDeleteHostByID(w http.ResponseWriter, r *http.Request) error {
	ctx, cancel := s.ServerContext(r.Context())
	defer cancel()

	id := chi.URLParam(r, "hostID")
	if id == "" {
		return echosight.ErrInvalidf("no host ID provided")
	}

	hostID, err := uuid.Parse(id)
	if err != nil {
		s.log.Errorc("invalid id: %v", err)
		return echosight.ErrInvalidf("invalid ID '%s'", id)
	}

	host, err := s.HostService.DeleteByID(ctx, hostID)
	if err != nil {
		return err
	}

	return writeJSON(w, http.StatusOK, Response{
		Status: StatusOK,
		Data: W{
			"host": host,
		},
	})
}
