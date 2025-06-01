package main

import (
	"log"
	"strconv"
	"sync"

	"github.com/metalmatze/transmission-exporter"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	namespace string = "transmission_"
)

// TorrentCollector has a transmission.Client to create torrent metrics
type TorrentCollector struct {
	client *transmission.Client

	Status             *prometheus.Desc
	Added              *prometheus.Desc
	Files              *prometheus.Desc
	Finished           *prometheus.Desc
	Done               *prometheus.Desc
	Ratio              *prometheus.Desc
	Download           *prometheus.Desc
	Upload             *prometheus.Desc
	PeersConnected     *prometheus.Desc
	PeersGettingFromUs *prometheus.Desc
	PeersSendingToUs   *prometheus.Desc
	TotalSize          *prometheus.Desc
	DownloadEver       *prometheus.Desc
	UploadedEver       *prometheus.Desc

	// TrackerStats
	Downloads *prometheus.Desc
	Leechers  *prometheus.Desc
	Seeders   *prometheus.Desc

	// Cache
	recentlyActiveOnly bool
	cachedTorrents     map[string]transmission.Torrent
	cachedTorrentsLock sync.Mutex
}

// NewTorrentCollector creates a new torrent collector with the transmission.Client
func NewTorrentCollector(client *transmission.Client) *TorrentCollector {
	const collectorNamespace = "torrent_"

	return &TorrentCollector{
		client:         client,
		cachedTorrents: make(map[string]transmission.Torrent),

		Status: prometheus.NewDesc(
			namespace+collectorNamespace+"status",
			"Status of a torrent",
			[]string{"id", "name"},
			nil,
		),
		Added: prometheus.NewDesc(
			namespace+collectorNamespace+"added",
			"The unixtime time a torrent was added",
			[]string{"id", "name"},
			nil,
		),
		Files: prometheus.NewDesc(
			namespace+collectorNamespace+"files_total",
			"The total number of files in a torrent",
			[]string{"id", "name"},
			nil,
		),
		Finished: prometheus.NewDesc(
			namespace+collectorNamespace+"finished",
			"Indicates if a torrent is finished (1) or not (0)",
			[]string{"id", "name"},
			nil,
		),
		Done: prometheus.NewDesc(
			namespace+collectorNamespace+"done",
			"The percent of a torrent being done",
			[]string{"id", "name"},
			nil,
		),
		Ratio: prometheus.NewDesc(
			namespace+collectorNamespace+"ratio",
			"The upload ratio of a torrent",
			[]string{"id", "name"},
			nil,
		),
		Download: prometheus.NewDesc(
			namespace+collectorNamespace+"download_bytes",
			"The current download rate of a torrent in bytes",
			[]string{"id", "name"},
			nil,
		),
		Upload: prometheus.NewDesc(
			namespace+collectorNamespace+"upload_bytes",
			"The current upload rate of a torrent in bytes",
			[]string{"id", "name"},
			nil,
		),
		PeersConnected: prometheus.NewDesc(
			namespace+collectorNamespace+"peers_connected",
			"The current number of peers connected to us",
			[]string{"id", "name"},
			nil,
		),
		PeersGettingFromUs: prometheus.NewDesc(
			namespace+collectorNamespace+"peers_getting_from_us",
			"The current number of peers downloading from us",
			[]string{"id", "name"},
			nil,
		),
		PeersSendingToUs: prometheus.NewDesc(
			namespace+collectorNamespace+"peers_sending_to_us",
			"The current number of peers sending to us",
			[]string{"id", "name"},
			nil,
		),
		TotalSize: prometheus.NewDesc(
			namespace+collectorNamespace+"total_size",
			"The total size of the torrent",
			[]string{"id", "name"},
			nil,
		),
		DownloadEver: prometheus.NewDesc(
			namespace+collectorNamespace+"downloaded_ever",
			"The total downloaded of the torrent",
			[]string{"id", "name"},
			nil,
		),
		UploadedEver: prometheus.NewDesc(
			namespace+collectorNamespace+"uploaded_ever",
			"The total uploaded of the torrent",
			[]string{"id", "name"},
			nil,
		),

		// TrackerStats
		Downloads: prometheus.NewDesc(
			namespace+collectorNamespace+"downloads_total",
			"How often this torrent was downloaded",
			[]string{"id", "name", "tracker"},
			nil,
		),
		Leechers: prometheus.NewDesc(
			namespace+collectorNamespace+"leechers",
			"The number of peers downloading this torrent",
			[]string{"id", "name", "tracker"},
			nil,
		),
		Seeders: prometheus.NewDesc(
			namespace+collectorNamespace+"seeders",
			"The number of peers uploading this torrent",
			[]string{"id", "name", "tracker"},
			nil,
		),
	}
}

// Describe implements the prometheus.Collector interface
func (tc *TorrentCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- tc.Status
	ch <- tc.Added
	ch <- tc.Files
	ch <- tc.Finished
	ch <- tc.Done
	ch <- tc.Ratio
	ch <- tc.Download
	ch <- tc.Upload
	ch <- tc.Downloads
	ch <- tc.Leechers
	ch <- tc.Seeders
	ch <- tc.PeersConnected
	ch <- tc.PeersGettingFromUs
	ch <- tc.PeersSendingToUs
	ch <- tc.TotalSize
	ch <- tc.DownloadEver
	ch <- tc.UploadedEver
}

// Collect implements the prometheus.Collector interface
func (tc *TorrentCollector) Collect(ch chan<- prometheus.Metric) {
	torrents, err := tc.client.GetTorrents(tc.recentlyActiveOnly)
	if err != nil {
		log.Printf("failed to get torrents: %v", err)
		return
	}

	tc.cachedTorrentsLock.Lock()
	var torrentsToUpdate []transmission.Torrent
	for _, t := range torrents {
		tc.cachedTorrents[t.HashString] = t
	}
	for _, t := range tc.cachedTorrents {
		torrentsToUpdate = append(torrentsToUpdate, t)
	}
	tc.cachedTorrentsLock.Unlock()

	if len(torrentsToUpdate) > 0 {
		tc.recentlyActiveOnly = true // only do this if successful
	}

	for _, t := range torrentsToUpdate {
		var finished float64

		id := strconv.Itoa(t.ID)

		if t.IsFinished {
			finished = 1
		}

		ch <- prometheus.MustNewConstMetric(
			tc.Status,
			prometheus.GaugeValue,
			float64(t.Status),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Added,
			prometheus.GaugeValue,
			float64(t.Added),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Files,
			prometheus.GaugeValue,
			float64(len(t.Files)),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Finished,
			prometheus.GaugeValue,
			finished,
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Done,
			prometheus.GaugeValue,
			t.PercentDone,
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Ratio,
			prometheus.GaugeValue,
			t.UploadRatio,
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Download,
			prometheus.GaugeValue,
			float64(t.RateDownload),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Upload,
			prometheus.GaugeValue,
			float64(t.RateUpload),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.PeersConnected,
			prometheus.GaugeValue,
			float64(t.PeersConnected),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.PeersGettingFromUs,
			prometheus.GaugeValue,
			float64(t.PeersGettingFromUs),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.PeersSendingToUs,
			prometheus.GaugeValue,
			float64(t.PeersSendingToUs),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.TotalSize,
			prometheus.GaugeValue,
			float64(t.TotalSize),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.DownloadEver,
			prometheus.GaugeValue,
			float64(t.DownloadEver),
			id, t.Name,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.UploadedEver,
			prometheus.GaugeValue,
			float64(t.UploadedEver),
			id, t.Name,
		)

		tstats := make(map[string]transmission.TrackerStat)

		for _, tracker := range t.TrackerStats {
			if tr, exists := tstats[tracker.Host]; exists {
				tr.DownloadCount += tracker.DownloadCount
			} else {
				tstats[tracker.Host] = tracker
			}
		}

		for _, tracker := range tstats {
			ch <- prometheus.MustNewConstMetric(
				tc.Downloads,
				prometheus.GaugeValue,
				float64(tracker.DownloadCount),
				id, t.Name, tracker.Host,
			)

			ch <- prometheus.MustNewConstMetric(
				tc.Leechers,
				prometheus.GaugeValue,
				float64(tracker.LeecherCount),
				id, t.Name, tracker.Host,
			)

			ch <- prometheus.MustNewConstMetric(
				tc.Seeders,
				prometheus.GaugeValue,
				float64(tracker.SeederCount),
				id, t.Name, tracker.Host,
			)
		}
	}
}
