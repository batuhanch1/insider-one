package prometheus

import "github.com/prometheus/client_golang/prometheus"

const (
	Consumer_Name_MessageTotal          = "consumer_messages_total"
	Consumer_Help_MessageTotal          = "Total consumed messages"
	Consumer_Name_ProcessDurationSecond = "consumer_processing_duration_seconds"
	Consumer_Help_ProcessDurationSecond = "Message processing duration"

	Http_Name_MessageTotal          = "http_requests_total"
	Http_Help_MessageTotal          = "Total number of HTTP requests"
	Http_Name_RequestDurationSecond = "http_request_duration_seconds"
	Http_Help_RequestDurationSecond = "HTTP request duration"
)

const (
	Label_Queue  = "queue"
	Label_Status = "status"
	Label_Method = "method"
	Label_Path   = "path"
)

type Prometheus struct {
	ConsumerMessagesTotal      *prometheus.CounterVec
	ConsumerProcessingDuration *prometheus.HistogramVec
	HttpRequestsTotal          *prometheus.CounterVec
	HttpRequestDuration        *prometheus.HistogramVec
}

func InitForConsumer() *Prometheus {
	wrapper := &Prometheus{
		ConsumerMessagesTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: Consumer_Name_MessageTotal,
			Help: Consumer_Help_MessageTotal,
		}, []string{Label_Queue, Label_Status}),
		ConsumerProcessingDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    Consumer_Name_ProcessDurationSecond,
			Help:    Consumer_Help_ProcessDurationSecond,
			Buckets: prometheus.DefBuckets,
		}, []string{Label_Queue}),
	}
	prometheus.MustRegister(wrapper.ConsumerProcessingDuration)
	prometheus.MustRegister(wrapper.ConsumerMessagesTotal)
	return wrapper
}

func InitForAPI() *Prometheus {
	wrapper := &Prometheus{
		HttpRequestsTotal: prometheus.NewCounterVec(prometheus.CounterOpts{
			Name: Http_Name_MessageTotal,
			Help: Http_Help_MessageTotal,
		},
			[]string{Label_Method, Label_Path, Label_Status}),
		HttpRequestDuration: prometheus.NewHistogramVec(prometheus.HistogramOpts{
			Name:    Http_Name_RequestDurationSecond,
			Help:    Http_Help_RequestDurationSecond,
			Buckets: prometheus.DefBuckets,
		},
			[]string{Label_Method, Label_Path}),
	}
	prometheus.MustRegister(wrapper.HttpRequestDuration)
	prometheus.MustRegister(wrapper.HttpRequestsTotal)
	return wrapper
}
