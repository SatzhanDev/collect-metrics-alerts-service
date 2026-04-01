package handler

import (
	"html/template"
	"net/http"
)

func (h *MetricsHandler) GetList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		w.Header().Set("Allow", http.MethodGet)
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

	gauges, counters, err := h.svc.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	tmpl := `
	<html>
	<body>
	<h1>Metrics</h1>

	<h2>Gauges</h2>
	<ul>
	{{range $key, $value := .Gauges}}
		<li>{{$key}} = {{$value}}</li>
	{{end}}
	</ul>

	<h2>Counters</h2>
	<ul>
	{{range $key, $value := .Counters}}
		<li>{{$key}} = {{$value}}</li>
	{{end}}
	</ul>

	</body>
	</html>
	`

	t := template.Must(template.New("metrics").Parse(tmpl))

	data := struct {
		Gauges   map[string]float64
		Counters map[string]int64
	}{
		Gauges:   gauges,
		Counters: counters,
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	err = t.Execute(w, data)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

}
