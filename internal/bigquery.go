package internal

import (
	"context"
	"time"

	"cloud.google.com/go/bigquery"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"

	"github.com/prometheus/alertmanager/template"

	"github.com/beeper/alertmanager-to-bigquery/internal/config"
)

const BigQueryInsertScope = "https://www.googleapis.com/auth/bigquery.insertdata"

type BigQueryAlert map[string]bigquery.Value

func (c BigQueryAlert) Save() (row map[string]bigquery.Value, insertID string, err error) {
	return c, "", nil
}

func getBigQueryTableInserter(cfg config.BigQueryConfig) (*bigquery.Inserter, error) {
	creds, err := google.CredentialsFromJSON(
		context.Background(),
		[]byte(cfg.CredentialsJSON), bigquery.Scope, BigQueryInsertScope,
	)
	if err != nil {
		return nil, err
	}

	client, err := bigquery.NewClient(context.Background(), cfg.ProjectID, option.WithCredentials(creds))
	if err != nil {
		return nil, err
	}

	dataset := client.Dataset(cfg.Dataset)
	_, err = dataset.Metadata(context.Background())
	if err != nil {
		return nil, err
	}

	table := dataset.Table(cfg.Table)
	_, err = table.Metadata(context.Background())
	if err != nil {
		return nil, err
	}

	return table.Inserter(), nil
}

func alertToBigQueryAlert(labelMap map[string]string, alert template.Alert) BigQueryAlert {
	bqAlert := map[string]bigquery.Value{
		"alertname":  alert.Labels["alertname"],
		"status":     alert.Status,
		"created_at": time.Now(),
	}

	for alertKey, bigQueryKey := range labelMap {
		value, ok := alert.Labels[alertKey]
		if ok {
			bqAlert[bigQueryKey] = value
		}
	}

	return bqAlert
}

func alertsToBigQueryAlerts(labelMap map[string]string, alerts []template.Alert) []BigQueryAlert {
	bqAlerts := make([]BigQueryAlert, 0, len(alerts))
	for _, alert := range alerts {
		bqAlerts = append(bqAlerts, alertToBigQueryAlert(labelMap, alert))
	}
	return bqAlerts
}
