package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/prometheus/client_golang/prometheus"
)

type PipelineMonitor struct {
	PipelineName string `json:"pipeline_name"`
	Status       string `json:"status"`
	Stages      []Stage `json:"stages"`
}

type Stage struct {
	StageName string `json:"stage_name"`
	Status    string `json:"status"`
}

var (
	pipelineMonitor = &PipelineMonitor{
		PipelineName: "My DevOps Pipeline",
		Status:       "Running",
		Stages: []Stage{
			{
				StageName: "Build",
				Status:    "Success",
			},
			{
				StageName: "Test",
				Status:    "Running",
			},
			{
				StageName: "Deploy",
				Status:    "Pending",
			},
		},
	}

	prometheusRegistry = prometheus.NewRegistry()
	pipelineMonitorGauge = prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "pipeline_monitor",
			Help: "DevOps pipeline monitor",
		},
		[]string{"pipeline_name", "stage_name", "status"},
	)
)

func init() {
	prometheusRegistry.MustRegister(pipelineMonitorGauge)
}

func getPipelineMonitor(w http.ResponseWriter, r *http.Request) {
	pipelineMonitorGauge.WithLabelValues(pipelineMonitor.PipelineName, "overall", pipelineMonitor.Status).Set(1)
	for _, stage := range pipelineMonitor.Stages {
		pipelineMonitorGauge.WithLabelValues(pipelineMonitor.PipelineName, stage.StageName, stage.Status).Set(1)
	}

	json.NewEncoder(w).Encode(pipelineMonitor)
}

func index(w http.ResponseWriter, r *http.Request) {
	t, _ := template.New("index").Parse(`
		<html>
			<head>
				<title>DevOps Pipeline Monitor</title>
			</head>
			<body>
				<h1>{{.PipelineName}} ({{.Status}})</h1>
				<ul>
					{{range .Stages}}
						<li>{{.StageName}} ({{.Status}})</li>
					{{end}}
				</ul>
			</body>
		</html>
	`)
	t.Execute(w, pipelineMonitor)
}

func main() {
	r := mux.NewRouter()
	r.HandleFunc("/pipeline", getPipelineMonitor).Methods("GET")
	r.HandleFunc("/", index).Methods("GET")

	http.Handle("/metrics", promhttp.HandlerFor(prometheusRegistry, promhttp.HandlerOpts{}))

	fmt.Println("Listening on port 8080")
	http.ListenAndServe(":8080", r)
}