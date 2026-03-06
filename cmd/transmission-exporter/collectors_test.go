package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	transmission "github.com/metalmatze/transmission-exporter"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

// mockTransmissionServer creates an httptest.Server that responds to Transmission RPC calls.
func mockTransmissionServer(t *testing.T) *httptest.Server {
	t.Helper()
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Transmission-Session-Id", "test-token")

		if r.Method != http.MethodPost {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}

		var cmd struct {
			Method string `json:"method"`
		}
		if err := json.NewDecoder(r.Body).Decode(&cmd); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		switch cmd.Method {
		case "torrent-get":
			json.NewEncoder(w).Encode(map[string]any{
				"result": "success",
				"arguments": map[string]any{
					"torrents": []map[string]any{
						{
							"id":                 1,
							"name":               "Test Torrent",
							"hashString":         "abc123def456",
							"status":             4,
							"addedDate":          1700000000,
							"activityDate":       1700001000,
							"doneDate":           0,
							"leftUntilDone":      1048576,
							"eta":                3600,
							"uploadRatio":        1.5,
							"rateDownload":       102400,
							"rateUpload":         51200,
							"downloadDir":        "/downloads",
							"isFinished":         false,
							"percentDone":        0.75,
							"seedRatioMode":      0,
							"error":              0,
							"errorString":        "",
							"files":              []map[string]any{{"name": "file1.mkv", "length": 4194304, "bytesCompleted": 3145728}},
							"fileStats":          []map[string]any{{"bytesCompleted": 3145728, "priority": 0, "wanted": true}},
							"trackerStats":       []map[string]any{{"host": "tracker.example.com", "isBackup": false, "downloadCount": 100, "leecherCount": 5, "seederCount": 20}},
							"peers":              []map[string]any{},
							"peersConnected":     10,
							"peersGettingFromUs": 3,
							"peersSendingToUs":   7,
							"totalSize":          4194304,
							"sizeWhenDone":       4194304,
							"downloadedEver":     3145728,
							"uploadedEver":       4718592,
							"corruptEver":        0,
							"haveValid":          3145728,
							"desiredAvailable":   1048576,
							"secondsDownloading": 7200,
							"secondsSeeding":     0,
							"queuePosition":      0,
							"pieceCount":         64,
							"pieceSize":          65536,
						},
						{
							"id":                 2,
							"name":               "Completed Torrent",
							"hashString":         "xyz789",
							"status":             6,
							"addedDate":          1699000000,
							"activityDate":       1700000500,
							"doneDate":           1699500000,
							"leftUntilDone":      0,
							"eta":                -1,
							"uploadRatio":        2.3,
							"rateDownload":       0,
							"rateUpload":         25600,
							"downloadDir":        "/downloads",
							"isFinished":         true,
							"percentDone":        1.0,
							"seedRatioMode":      0,
							"error":              0,
							"errorString":        "",
							"files":              []map[string]any{{"name": "movie.mp4", "length": 2097152, "bytesCompleted": 2097152}},
							"fileStats":          []map[string]any{{"bytesCompleted": 2097152, "priority": 0, "wanted": true}},
							"trackerStats":       []map[string]any{{"host": "tracker2.example.com", "isBackup": false, "downloadCount": 50, "leecherCount": 2, "seederCount": 30}},
							"peers":              []map[string]any{},
							"peersConnected":     5,
							"peersGettingFromUs": 4,
							"peersSendingToUs":   0,
							"totalSize":          2097152,
							"sizeWhenDone":       2097152,
							"downloadedEver":     2097152,
							"uploadedEver":       4823449,
							"corruptEver":        1024,
							"haveValid":          2097152,
							"desiredAvailable":   0,
							"secondsDownloading": 3600,
							"secondsSeeding":     86400,
							"queuePosition":      1,
							"pieceCount":         32,
							"pieceSize":          65536,
						},
					},
				},
			})
		case "session-get":
			json.NewEncoder(w).Encode(map[string]any{
				"result": "success",
				"arguments": map[string]any{
					"alt-speed-down":         100,
					"alt-speed-enabled":      false,
					"alt-speed-up":           50,
					"cache-size-mb":          4,
					"download-dir":           "/downloads",
					"download-dir-free-space": 107374182400,
					"download-queue-enabled": true,
					"download-queue-size":    5,
					"incomplete-dir":         "/incomplete",
					"peer-limit-global":      200,
					"peer-limit-per-torrent": 50,
					"seed-queue-enabled":     false,
					"seed-queue-size":        10,
					"seedRatioLimit":         2.0,
					"seedRatioLimited":       true,
					"speed-limit-down":       1000,
					"speed-limit-down-enabled": false,
					"speed-limit-up":         500,
					"speed-limit-up-enabled": false,
					"version":               "4.0.6",
				},
			})
		case "session-stats":
			json.NewEncoder(w).Encode(map[string]any{
				"result": "success",
				"arguments": map[string]any{
					"downloadSpeed":      524288,
					"uploadSpeed":        262144,
					"activeTorrentCount": 1,
					"pausedTorrentCount": 1,
					"torrentCount":       2,
					"current-stats": map[string]any{
						"downloadedBytes": 1073741824,
						"uploadedBytes":   536870912,
						"filesAdded":      5,
						"secondsActive":   86400,
						"sessionCount":    1,
					},
					"cumulative-stats": map[string]any{
						"downloadedBytes": 10737418240,
						"uploadedBytes":   5368709120,
						"filesAdded":      50,
						"secondsActive":   864000,
						"sessionCount":    10,
					},
				},
			})
		default:
			w.WriteHeader(http.StatusBadRequest)
		}
	}))
}

func newTestClient(t *testing.T) (*transmission.Client, *httptest.Server) {
	t.Helper()
	srv := mockTransmissionServer(t)
	client := transmission.New(srv.URL, nil)
	return client, srv
}

func TestTorrentCollector(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	collector := NewTorrentCollector(client)
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)

	// Gather all metrics
	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	// Build a map of metric name -> metric family
	metrics := make(map[string]*struct {
		count int
		help  string
	})
	for _, f := range families {
		metrics[f.GetName()] = &struct {
			count int
			help  string
		}{count: len(f.GetMetric()), help: f.GetHelp()}
	}

	// Verify key metrics exist with expected counts (2 torrents)
	expectedMetrics := []string{
		"transmission_torrent_status",
		"transmission_torrent_added",
		"transmission_torrent_done",
		"transmission_torrent_ratio",
		"transmission_torrent_download_bytes",
		"transmission_torrent_upload_bytes",
		"transmission_torrent_peers_connected",
		"transmission_torrent_total_size",
		"transmission_torrent_downloaded_ever",
		"transmission_torrent_uploaded_ever",
		"transmission_torrent_size_when_done_bytes",
		"transmission_torrent_corrupt_ever_bytes",
		"transmission_torrent_have_valid_bytes",
		"transmission_torrent_desired_available_bytes",
		"transmission_torrent_seconds_downloading",
		"transmission_torrent_seconds_seeding",
		"transmission_torrent_queue_position",
		"transmission_torrent_activity_date",
		"transmission_torrent_done_date",
		"transmission_torrent_error",
		"transmission_torrent_left_until_done_bytes",
	}

	for _, name := range expectedMetrics {
		m, ok := metrics[name]
		if !ok {
			t.Errorf("missing metric %s", name)
			continue
		}
		if m.count != 2 {
			t.Errorf("metric %s: expected 2 samples (2 torrents), got %d", name, m.count)
		}
	}

	// Verify specific values using testutil
	expected := `
		# HELP transmission_torrent_done The percent of a torrent being done
		# TYPE transmission_torrent_done gauge
		transmission_torrent_done{id="1",name="Test Torrent",tracker="tracker.example.com"} 0.75
		transmission_torrent_done{id="2",name="Completed Torrent",tracker="tracker2.example.com"} 1
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "transmission_torrent_done"); err != nil {
		t.Errorf("torrent_done mismatch: %v", err)
	}
}

func TestSessionCollector(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	collector := NewSessionCollector(client)
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	metricNames := make(map[string]bool)
	for _, f := range families {
		metricNames[f.GetName()] = true
	}

	expectedMetrics := []string{
		"transmission_alt_speed_down",
		"transmission_alt_speed_up",
		"transmission_cache_size_bytes",
		"transmission_free_space",
		"transmission_queue_down",
		"transmission_queue_up",
		"transmission_global_peer_limit",
		"transmission_torrent_peer_limit",
		"transmission_seed_ratio_limit",
		"transmission_speed_limit_down_bytes",
		"transmission_speed_limit_up_bytes",
		"transmission_version",
	}

	for _, name := range expectedMetrics {
		if !metricNames[name] {
			t.Errorf("missing session metric %s", name)
		}
	}

	// Check specific values
	expected := `
		# HELP transmission_global_peer_limit Maximum global number of peers
		# TYPE transmission_global_peer_limit gauge
		transmission_global_peer_limit 200
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "transmission_global_peer_limit"); err != nil {
		t.Errorf("global_peer_limit mismatch: %v", err)
	}

	expected = `
		# HELP transmission_version Transmission version as label
		# TYPE transmission_version gauge
		transmission_version{version="4.0.6"} 1
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "transmission_version"); err != nil {
		t.Errorf("version mismatch: %v", err)
	}
}

func TestSessionStatsCollector(t *testing.T) {
	client, srv := newTestClient(t)
	defer srv.Close()

	collector := NewSessionStatsCollector(client)
	reg := prometheus.NewPedanticRegistry()
	reg.MustRegister(collector)

	families, err := reg.Gather()
	if err != nil {
		t.Fatalf("failed to gather metrics: %v", err)
	}

	metricNames := make(map[string]bool)
	for _, f := range families {
		metricNames[f.GetName()] = true
	}

	expectedMetrics := []string{
		"transmission_session_stats_download_speed_bytes",
		"transmission_session_stats_upload_speed_bytes",
		"transmission_session_stats_torrents_total",
		"transmission_session_stats_torrents_active",
		"transmission_session_stats_torrents_paused",
		"transmission_session_stats_downloaded_bytes",
		"transmission_session_stats_uploaded_bytes",
		"transmission_session_stats_files_added",
		"transmission_session_stats_sessions",
	}

	for _, name := range expectedMetrics {
		if !metricNames[name] {
			t.Errorf("missing session stats metric %s", name)
		}
	}

	// Check specific values
	expected := `
		# HELP transmission_session_stats_torrents_total The total number of torrents
		# TYPE transmission_session_stats_torrents_total gauge
		transmission_session_stats_torrents_total 2
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "transmission_session_stats_torrents_total"); err != nil {
		t.Errorf("torrents_total mismatch: %v", err)
	}
}

func TestTorrentCollectorErrorHandling(t *testing.T) {
	// Server that returns errors
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Transmission-Session-Id", "test-token")
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := transmission.New(srv.URL, nil)
	collector := NewTorrentCollector(client)

	// Should not panic on error
	ch := make(chan prometheus.Metric, 100)
	collector.Collect(ch)
	close(ch)

	count := 0
	for range ch {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 metrics on error, got %d", count)
	}
}

func TestTorrentCollectorNilResponse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Transmission-Session-Id", "test-token")
		json.NewEncoder(w).Encode(map[string]any{
			"result":    "success",
			"arguments": map[string]any{},
		})
	}))
	defer srv.Close()

	client := transmission.New(srv.URL, nil)
	collector := NewTorrentCollector(client)

	ch := make(chan prometheus.Metric, 100)
	collector.Collect(ch)
	close(ch)

	count := 0
	for range ch {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 metrics for nil torrents, got %d", count)
	}
}

func TestTorrentCachePruning(t *testing.T) {
	scrapeCount := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Transmission-Session-Id", "test-token")

		var cmd struct{ Method string `json:"method"` }
		json.NewDecoder(r.Body).Decode(&cmd)

		if cmd.Method != "torrent-get" {
			w.WriteHeader(http.StatusBadRequest)
			return
		}

		scrapeCount++
		switch {
		case scrapeCount == 1:
			// First scrape: two torrents
			json.NewEncoder(w).Encode(map[string]any{
				"result": "success",
				"arguments": map[string]any{
					"torrents": []map[string]any{
						torrentFixture("hash1", "Torrent A"),
						torrentFixture("hash2", "Torrent B"),
					},
				},
			})
		case scrapeCount > 1 && scrapeCount < 10:
			// Scrapes 2-9: recently-active returns only Torrent A
			json.NewEncoder(w).Encode(map[string]any{
				"result": "success",
				"arguments": map[string]any{
					"torrents": []map[string]any{
						torrentFixture("hash1", "Torrent A"),
					},
				},
			})
		case scrapeCount == 10:
			// Scrape 10: full refresh — Torrent B was removed
			json.NewEncoder(w).Encode(map[string]any{
				"result": "success",
				"arguments": map[string]any{
					"torrents": []map[string]any{
						torrentFixture("hash1", "Torrent A"),
					},
				},
			})
		default:
			json.NewEncoder(w).Encode(map[string]any{
				"result": "success",
				"arguments": map[string]any{
					"torrents": []map[string]any{
						torrentFixture("hash1", "Torrent A"),
					},
				},
			})
		}
	}))
	defer srv.Close()

	client := transmission.New(srv.URL, nil)
	collector := NewTorrentCollector(client)

	collectMetricCount := func() int {
		ch := make(chan prometheus.Metric, 1000)
		collector.Collect(ch)
		close(ch)
		count := 0
		for range ch {
			count++
		}
		return count
	}

	// Scrape 1 (full): should have 2 torrents
	count1 := collectMetricCount()
	if count1 == 0 {
		t.Fatal("expected metrics from first scrape")
	}

	// Scrapes 2-9 (recently-active only): Torrent B stays in cache
	var countMid int
	for i := 0; i < 8; i++ {
		countMid = collectMetricCount()
	}
	// Both torrents should still be cached even though only A was returned
	if countMid != count1 {
		t.Errorf("expected cached torrent count to remain %d during recently-active scrapes, got %d", count1, countMid)
	}

	// Scrape 10: full refresh, Torrent B should be pruned
	count10 := collectMetricCount()
	// Should have fewer metrics now (only 1 torrent instead of 2)
	if count10 >= count1 {
		t.Errorf("expected fewer metrics after pruning (1 torrent), but got %d (was %d with 2 torrents)", count10, count1)
	}

	// Verify it's roughly half the torrent metrics (some tracker metrics may differ)
	// Each torrent emits ~27 base metrics + tracker metrics
	expectedRatio := float64(count10) / float64(count1)
	if expectedRatio > 0.6 {
		t.Errorf("expected ~50%% of metrics after pruning 1 of 2 torrents, got %.0f%%", expectedRatio*100)
	}
}

func torrentFixture(hash, name string) map[string]any {
	return map[string]any{
		"id": 1, "name": name, "hashString": hash,
		"status": 4, "addedDate": 1700000000, "activityDate": 1700001000,
		"doneDate": 0, "leftUntilDone": 0, "eta": -1,
		"uploadRatio": 1.0, "rateDownload": 0, "rateUpload": 0,
		"downloadDir": "/dl", "isFinished": true, "percentDone": 1.0,
		"seedRatioMode": 0, "error": 0, "errorString": "",
		"files": []any{}, "fileStats": []any{},
		"trackerStats": []map[string]any{
			{"host": "tracker.test.com", "isBackup": false, "downloadCount": 10, "leecherCount": 1, "seederCount": 5},
		},
		"peers": []any{}, "peersConnected": 0, "peersGettingFromUs": 0,
		"peersSendingToUs": 0, "totalSize": 100, "sizeWhenDone": 100,
		"downloadedEver": 100, "uploadedEver": 100,
		"corruptEver": 0, "haveValid": 100, "desiredAvailable": 0,
		"secondsDownloading": 60, "secondsSeeding": 120,
		"queuePosition": 0, "pieceCount": 1, "pieceSize": 100,
	}
}

func TestTrackerStatsAggregation(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Transmission-Session-Id", "test-token")

		var cmd struct{ Method string `json:"method"` }
		json.NewDecoder(r.Body).Decode(&cmd)

		if cmd.Method == "torrent-get" {
			json.NewEncoder(w).Encode(map[string]any{
				"result": "success",
				"arguments": map[string]any{
					"torrents": []map[string]any{
						{
							"id": 1, "name": "Multi-Tracker", "hashString": "multi123",
							"status": 4, "addedDate": 1700000000, "activityDate": 1700001000,
							"doneDate": 0, "leftUntilDone": 0, "eta": -1,
							"uploadRatio": 1.0, "rateDownload": 0, "rateUpload": 0,
							"downloadDir": "/dl", "isFinished": true, "percentDone": 1.0,
							"seedRatioMode": 0, "error": 0, "errorString": "",
							"files": []any{}, "fileStats": []any{},
							"trackerStats": []map[string]any{
								{"host": "tracker.a.com", "isBackup": false, "downloadCount": 10, "leecherCount": 1, "seederCount": 5},
								{"host": "tracker.a.com", "isBackup": true, "downloadCount": 20, "leecherCount": 2, "seederCount": 10},
							},
							"peers": []any{}, "peersConnected": 0, "peersGettingFromUs": 0,
							"peersSendingToUs": 0, "totalSize": 100, "sizeWhenDone": 100,
							"downloadedEver": 100, "uploadedEver": 100,
							"corruptEver": 0, "haveValid": 100, "desiredAvailable": 0,
							"secondsDownloading": 60, "secondsSeeding": 120,
							"queuePosition": 0, "pieceCount": 1, "pieceSize": 100,
						},
					},
				},
			})
		}
	}))
	defer srv.Close()

	client := transmission.New(srv.URL, nil)
	collector := NewTorrentCollector(client)

	// The two tracker stats for the same host should be aggregated
	expected := `
		# HELP transmission_torrent_downloads_total How often this torrent was downloaded
		# TYPE transmission_torrent_downloads_total gauge
		transmission_torrent_downloads_total{id="1",name="Multi-Tracker",tracker="tracker.a.com"} 30
	`
	if err := testutil.CollectAndCompare(collector, strings.NewReader(expected), "transmission_torrent_downloads_total"); err != nil {
		t.Errorf("tracker aggregation mismatch: %v", err)
	}
}
