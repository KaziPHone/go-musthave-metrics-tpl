package storage

import (
	"testing"
)

func TestMStorage_UpdateMetric(t *testing.T) {
	type fields struct {
		MetricTypes map[string]*MetricType
	}
	type args struct {
		metricName string
		typeMetric string
		value      interface{}
	}
	tests := []struct {
		name   string
		fields fields
		args   args
	}{
		{
			name: "test gauge",
			fields: fields{
				MetricTypes: map[string]*MetricType{
					"metricName": &MetricType{
						Gauge:   0,
						Counter: 0,
					},
				},
			},
			args: args{
				metricName: "metricName gauge",
				typeMetric: "gauge",
				value:      100.0,
			},
		},
		{
			name: "test counter",
			fields: fields{
				MetricTypes: map[string]*MetricType{
					"metricName": &MetricType{
						Gauge:   0,
						Counter: 0,
					},
				},
			},
			args: args{
				metricName: "metricName counter",
				typeMetric: "counter",
				value:      int64(100),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := NewMemStorage()
			m.UpdateMetric(tt.args.metricName, tt.args.typeMetric, tt.args.value)
			v := m.ListMetrics()
			if tt.args.typeMetric == "gauge" {

				if v[tt.args.metricName].Gauge != tt.args.value {
					t.Errorf(
						"expected %v, got %v",
						tt.args.value,
						v[tt.args.metricName].Gauge,
					)
				}
			} else {
				if v[tt.args.metricName].Counter != tt.args.value {
					t.Errorf(
						"expected %v, got %v",
						tt.args.value,
						v[tt.args.metricName].Counter,
					)
				}
			}

		})
	}
}
