package storage

import "testing"

func TestMemStorage_UpdateMetric(t *testing.T) {
	type fields struct {
		MetricTypes map[string]*metricType
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
			name: "test",
			fields: fields{
				MetricTypes: map[string]*metricType{
					"metricName": &metricType{
						Gauge:   0,
						Counter: 0,
					},
				},
			},
			args: args{
				metricName: "metricName",
				typeMetric: "gauge",
				value:      100.0,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MemStorage{
				MetricTypes: tt.fields.MetricTypes,
			}
			m.UpdateMetric(tt.args.metricName, tt.args.typeMetric, tt.args.value)
			if m.MetricTypes[tt.args.metricName].Gauge != tt.args.value {
				t.Errorf(
					"expected %v, got %v",
					tt.args.value,
					m.MetricTypes[tt.args.metricName].Gauge,
				)
			}
		})
	}
}
