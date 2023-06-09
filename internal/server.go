package internal

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"cloud.google.com/go/bigquery"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/alertmanager/template"
	"github.com/rs/zerolog/hlog"
	"github.com/rs/zerolog/log"

	"github.com/beeper/alertmanager-to-bigquery/internal/config"
	"github.com/beeper/libserv/pkg/health"
	"github.com/beeper/libserv/pkg/requestlog"
)

type AlertManagerToBigQuery struct {
	config   config.AlertmanagerToBigQueryConfig
	inserter *bigquery.Inserter
}

func NewAlertManagerToBigQuery(cfg config.AlertmanagerToBigQueryConfig) *AlertManagerToBigQuery {
	amtobq := AlertManagerToBigQuery{config: cfg}
	return &amtobq
}

func (amtobq *AlertManagerToBigQuery) Start() {
	inserter, err := getBigQueryTableInserter(amtobq.config.BigQuery)
	if err != nil {
		panic(err)
	}
	amtobq.inserter = inserter

	r := chi.NewRouter()
	r.Use(hlog.NewHandler(log.Logger))
	r.Use(hlog.RequestIDHandler("request_id", ""))
	r.Use(requestlog.AccessLogger(true))
	r.Use(middleware.Recoverer)

	routes := []*requestlog.Route{
		{Path: "/notification", Method: http.MethodPost, Handler: amtobq.handleNotification},
		{Path: "/health", Method: http.MethodGet, Handler: health.Health},
	}

	for _, rt := range routes {
		r.Method(rt.Method, rt.Path, rt)
	}

	server := &http.Server{
		Addr:    amtobq.config.Server.Host + ":" + strconv.Itoa(amtobq.config.Server.Port),
		Handler: r,
	}
	log.Info().Str("addr", server.Addr).Msg("starting listener")
	err = server.ListenAndServe()

	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().Err(err).Msg("error while listening")
	} else {
		log.Info().Msg("listener stopped")
	}
}

func (amtobq *AlertManagerToBigQuery) handleNotification(w http.ResponseWriter, r *http.Request) {
	var data template.Data

	if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
		log.Error().Err(err).Msg("Error while decoding request from alertmanager")
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	bqAlerts := alertsToBigQueryAlerts(amtobq.config.LabelMap, data.Alerts)
	err := amtobq.inserter.Put(r.Context(), bqAlerts)
	if err != nil {
		log.Error().Err(err).Msg("Failed to insert to BigQuery")
		w.WriteHeader(http.StatusServiceUnavailable)
		return
	}

	log.Info().
		Int("alerts", len(data.Alerts)).
		Msg("Alerts inserted to BigQuery")

	w.WriteHeader(http.StatusNoContent)
}
