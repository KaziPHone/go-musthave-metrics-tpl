package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/KaziPHone/go-musthave-metrics-tpl/pkg/storage"
	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
)

func TestHandler_GetMetricHandler(t *testing.T) {
	req := httptest.NewRequest("GET", "/value/gauge/test_metric", nil)
	respRec := httptest.NewRecorder()
	ctx := context.Background()
	routeCtx := chi.NewRouteContext()
	routeCtx.URLParams.Add("typeMetric", "gauge")
	routeCtx.URLParams.Add("nameMetric", "test_metric")
	req = req.WithContext(ctx)
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, routeCtx))
	type fields struct {
		Storage *storage.MStorage
	}
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		// TODO: Add test cases.
		{
			name: "test",
			fields: fields{
				Storage: &storage.MStorage{
					MetricTypes: map[string]*storage.MetricType{
						"test_metric": &storage.MetricType{
							Gauge:   100.0,
							Counter: 10,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &Handler{
				Storage: tt.fields.Storage,
			}
			h.GetMetricHandler(respRec, req)
			assert.Equal(t, respRec.Code, http.StatusOK, "expected status code to be 200")
		})
	}
}
