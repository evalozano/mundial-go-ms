package frontend

import (
	"encoding/json"
	"fmt"
	"net/http"

	profile "mundial-go-ms/services/profile/proto"
	search "mundial-go-ms/services/search/proto"
	"mundial-go-ms/tracing"
	opentracing "github.com/opentracing/opentracing-go"
)

// NewServer returns a new server
func NewServer(sc search.SearchClient, pc profile.ProfileClient, tr opentracing.Tracer) *Server {
	return &Server{
		searchClient:  sc,
		profileClient: pc,
		tracer:        tr,
	}
}

// Server implements frontend service
type Server struct {
	searchClient  search.SearchClient
	profileClient profile.ProfileClient
	tracer        opentracing.Tracer
}

// Run the server
func (s *Server) Run(port int) error {
	mux := tracing.NewServeMux(s.tracer)
	mux.Handle("/", http.FileServer(http.Dir("services/frontend/static")))

	// API
	mux.Handle("/pubs", http.HandlerFunc(s.searchHandler))

	return http.ListenAndServe(fmt.Sprintf(":%d", port), mux)
}

func (s *Server) searchHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	ctx := r.Context()

	// in/out dates from query params
	inDate, outDate := r.URL.Query().Get("inDate"), r.URL.Query().Get("outDate")
	if inDate == "" || outDate == "" {
		http.Error(w, "Please specify inDate/outDate params", http.StatusBadRequest)
		return
	}

	// search for best pubs
	// TODO(hw): allow lat/lon from input params
	searchResp, err := s.searchClient.Nearby(ctx, &search.NearbyRequest{
		Lat:     43.402259,
		Lon:     39.955290,
		InDate:  inDate,
		OutDate: outDate,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// grab locale from query params or default to en
	locale := r.URL.Query().Get("locale")
	if locale == "" {
		locale = "en"
	}

	// pub profiles
	profileResp, err := s.profileClient.GetProfiles(ctx, &profile.Request{
		PubIds: searchResp.PubIds,
		Locale:   locale,
	})
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(geoJSONResponse(profileResp.Pubs))
}

// return a geoJSON response that allows google map to plot points directly on map
// https://developers.google.com/maps/documentation/javascript/datalayer#sample_geojson
func geoJSONResponse(hs []*profile.Pub) map[string]interface{} {
	fs := []interface{}{}

	for _, h := range hs {
		fs = append(fs, map[string]interface{}{
			"type": "Feature",
			"id":   h.Id,
			"properties": map[string]string{
				"name":         h.Name,
				"phone_number": h.PhoneNumber,
			},
			"geometry": map[string]interface{}{
				"type": "Point",
				"coordinates": []float32{
					h.Address.Lon,
					h.Address.Lat,
				},
			},
		})
	}

	return map[string]interface{}{
		"type":     "FeatureCollection",
		"features": fs,
	}
}
