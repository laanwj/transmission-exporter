package main

import (
	"log/slog"
	"strconv"
	"sync"
	"time"

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

	SizeWhenDone       *prometheus.Desc
	CorruptEver        *prometheus.Desc
	HaveValid          *prometheus.Desc
	DesiredAvailable   *prometheus.Desc
	SecondsDownloading *prometheus.Desc
	SecondsSeeding     *prometheus.Desc
	QueuePosition      *prometheus.Desc
	ActivityDate       *prometheus.Desc
	DoneDate           *prometheus.Desc
	ErrorStatus        *prometheus.Desc
	LeftUntilDone      *prometheus.Desc

	// TrackerStats
	Downloads *prometheus.Desc
	Leechers  *prometheus.Desc
	Seeders   *prometheus.Desc

	// Cache
	recentlyActiveOnly bool
	scrapeCount        int
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
			[]string{"id", "name", "tracker"},
			nil,
		),
		Added: prometheus.NewDesc(
			namespace+collectorNamespace+"added",
			"The unixtime time a torrent was added",
			[]string{"id", "name", "tracker"},
			nil,
		),
		Files: prometheus.NewDesc(
			namespace+collectorNamespace+"files_total",
			"The total number of files in a torrent",
			[]string{"id", "name", "tracker"},
			nil,
		),
		Finished: prometheus.NewDesc(
			namespace+collectorNamespace+"finished",
			"Indicates if a torrent is finished (1) or not (0)",
			[]string{"id", "name", "tracker"},
			nil,
		),
		Done: prometheus.NewDesc(
			namespace+collectorNamespace+"done",
			"The percent of a torrent being done",
			[]string{"id", "name", "tracker"},
			nil,
		),
		Ratio: prometheus.NewDesc(
			namespace+collectorNamespace+"ratio",
			"The upload ratio of a torrent",
			[]string{"id", "name", "tracker"},
			nil,
		),
		Download: prometheus.NewDesc(
			namespace+collectorNamespace+"download_bytes",
			"The current download rate of a torrent in bytes",
			[]string{"id", "name", "tracker"},
			nil,
		),
		Upload: prometheus.NewDesc(
			namespace+collectorNamespace+"upload_bytes",
			"The current upload rate of a torrent in bytes",
			[]string{"id", "name", "tracker"},
			nil,
		),
		PeersConnected: prometheus.NewDesc(
			namespace+collectorNamespace+"peers_connected",
			"The current number of peers connected to us",
			[]string{"id", "name", "tracker"},
			nil,
		),
		PeersGettingFromUs: prometheus.NewDesc(
			namespace+collectorNamespace+"peers_getting_from_us",
			"The current number of peers downloading from us",
			[]string{"id", "name", "tracker"},
			nil,
		),
		PeersSendingToUs: prometheus.NewDesc(
			namespace+collectorNamespace+"peers_sending_to_us",
			"The current number of peers sending to us",
			[]string{"id", "name", "tracker"},
			nil,
		),
		TotalSize: prometheus.NewDesc(
			namespace+collectorNamespace+"total_size",
			"The total size of the torrent",
			[]string{"id", "name", "tracker"},
			nil,
		),
		DownloadEver: prometheus.NewDesc(
			namespace+collectorNamespace+"downloaded_ever",
			"The total downloaded of the torrent",
			[]string{"id", "name", "tracker"},
			nil,
		),
		UploadedEver: prometheus.NewDesc(
			namespace+collectorNamespace+"uploaded_ever",
			"The total uploaded of the torrent",
			[]string{"id", "name", "tracker"},
			nil,
		),
		SizeWhenDone: prometheus.NewDesc(
			namespace+collectorNamespace+"size_when_done_bytes",
			"Final size of the torrent when download is complete",
			[]string{"id", "name", "tracker"},
			nil,
		),
		CorruptEver: prometheus.NewDesc(
			namespace+collectorNamespace+"corrupt_ever_bytes",
			"Total corrupt bytes received for the torrent",
			[]string{"id", "name", "tracker"},
			nil,
		),
		HaveValid: prometheus.NewDesc(
			namespace+collectorNamespace+"have_valid_bytes",
			"Total verified bytes of the torrent",
			[]string{"id", "name", "tracker"},
			nil,
		),
		DesiredAvailable: prometheus.NewDesc(
			namespace+collectorNamespace+"desired_available_bytes",
			"Bytes we still need that are available from peers",
			[]string{"id", "name", "tracker"},
			nil,
		),
		SecondsDownloading: prometheus.NewDesc(
			namespace+collectorNamespace+"seconds_downloading",
			"Total seconds spent downloading the torrent",
			[]string{"id", "name", "tracker"},
			nil,
		),
		SecondsSeeding: prometheus.NewDesc(
			namespace+collectorNamespace+"seconds_seeding",
			"Total seconds spent seeding the torrent",
			[]string{"id", "name", "tracker"},
			nil,
		),
		QueuePosition: prometheus.NewDesc(
			namespace+collectorNamespace+"queue_position",
			"Position of the torrent in the download/seed queue",
			[]string{"id", "name", "tracker"},
			nil,
		),
		ActivityDate: prometheus.NewDesc(
			namespace+collectorNamespace+"activity_date",
			"Unix timestamp of the torrent's last activity",
			[]string{"id", "name", "tracker"},
			nil,
		),
		DoneDate: prometheus.NewDesc(
			namespace+collectorNamespace+"done_date",
			"Unix timestamp when the torrent finished downloading",
			[]string{"id", "name", "tracker"},
			nil,
		),
		ErrorStatus: prometheus.NewDesc(
			namespace+collectorNamespace+"error",
			"Torrent error status (0=OK, 1=tracker_warning, 2=tracker_error, 3=local_error)",
			[]string{"id", "name", "tracker"},
			nil,
		),
		LeftUntilDone: prometheus.NewDesc(
			namespace+collectorNamespace+"left_until_done_bytes",
			"Bytes remaining to download",
			[]string{"id", "name", "tracker"},
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
	ch <- tc.SizeWhenDone
	ch <- tc.CorruptEver
	ch <- tc.HaveValid
	ch <- tc.DesiredAvailable
	ch <- tc.SecondsDownloading
	ch <- tc.SecondsSeeding
	ch <- tc.QueuePosition
	ch <- tc.ActivityDate
	ch <- tc.DoneDate
	ch <- tc.ErrorStatus
	ch <- tc.LeftUntilDone
}

// Collect implements the prometheus.Collector interface
func (tc *TorrentCollector) Collect(ch chan<- prometheus.Metric) {
	start := time.Now()

	// Force a full refresh every 10 scrapes to prune removed torrents
	tc.cachedTorrentsLock.Lock()
	tc.scrapeCount++
	if tc.scrapeCount%10 == 0 {
		tc.recentlyActiveOnly = false
	}
	recentlyActiveOnly := tc.recentlyActiveOnly
	tc.cachedTorrentsLock.Unlock()

	torrents, err := tc.client.GetTorrents(recentlyActiveOnly)
	if err != nil {
		slog.Error("failed to get torrents", "error", err, "recently_active_only", tc.recentlyActiveOnly)
		return
	}

	if torrents == nil {
		slog.Warn("got nil torrents from transmission")
		return
	}

	tc.cachedTorrentsLock.Lock()
	if !recentlyActiveOnly {
		// Full fetch — replace the entire cache to prune removed torrents
		tc.cachedTorrents = make(map[string]transmission.Torrent, len(torrents))
	}
	for _, t := range torrents {
		tc.cachedTorrents[t.HashString] = t
	}
	torrentsToUpdate := make([]transmission.Torrent, 0, len(tc.cachedTorrents))
	for _, t := range tc.cachedTorrents {
		torrentsToUpdate = append(torrentsToUpdate, t)
	}
	if len(torrentsToUpdate) > 0 {
		tc.recentlyActiveOnly = true
	}
	tc.cachedTorrentsLock.Unlock()

	slog.Debug("collected torrents",
		"count", len(torrentsToUpdate),
		"fetched", len(torrents),
		"cached", len(tc.cachedTorrents),
		"duration", time.Since(start).String(),
	)

	for _, t := range torrentsToUpdate {
		var finished float64

		id := strconv.Itoa(t.ID)

		if t.IsFinished {
			finished = 1
		}

		// Use the first non-backup tracker as the primary tracker label
		tracker := ""
		for _, ts := range t.TrackerStats {
			if !ts.IsBackup {
				tracker = ts.Host
				break
			}
		}
		if tracker == "" && len(t.TrackerStats) > 0 {
			tracker = t.TrackerStats[0].Host
		}

		ch <- prometheus.MustNewConstMetric(
			tc.Status,
			prometheus.GaugeValue,
			float64(t.Status),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Added,
			prometheus.GaugeValue,
			float64(t.Added),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Files,
			prometheus.GaugeValue,
			float64(len(t.Files)),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Finished,
			prometheus.GaugeValue,
			finished,
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Done,
			prometheus.GaugeValue,
			t.PercentDone,
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Ratio,
			prometheus.GaugeValue,
			t.UploadRatio,
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Download,
			prometheus.GaugeValue,
			float64(t.RateDownload),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.Upload,
			prometheus.GaugeValue,
			float64(t.RateUpload),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.PeersConnected,
			prometheus.GaugeValue,
			float64(t.PeersConnected),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.PeersGettingFromUs,
			prometheus.GaugeValue,
			float64(t.PeersGettingFromUs),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.PeersSendingToUs,
			prometheus.GaugeValue,
			float64(t.PeersSendingToUs),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.TotalSize,
			prometheus.GaugeValue,
			float64(t.TotalSize),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.DownloadEver,
			prometheus.GaugeValue,
			float64(t.DownloadEver),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.UploadedEver,
			prometheus.GaugeValue,
			float64(t.UploadedEver),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.SizeWhenDone,
			prometheus.GaugeValue,
			float64(t.SizeWhenDone),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.CorruptEver,
			prometheus.GaugeValue,
			float64(t.CorruptEver),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.HaveValid,
			prometheus.GaugeValue,
			float64(t.HaveValid),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.DesiredAvailable,
			prometheus.GaugeValue,
			float64(t.DesiredAvailable),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.SecondsDownloading,
			prometheus.GaugeValue,
			float64(t.SecondsDownloading),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.SecondsSeeding,
			prometheus.GaugeValue,
			float64(t.SecondsSeeding),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.QueuePosition,
			prometheus.GaugeValue,
			float64(t.QueuePosition),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.ActivityDate,
			prometheus.GaugeValue,
			float64(t.ActivityDate),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.DoneDate,
			prometheus.GaugeValue,
			float64(t.DoneDate),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.ErrorStatus,
			prometheus.GaugeValue,
			float64(t.Error),
			id, t.Name, tracker,
		)
		ch <- prometheus.MustNewConstMetric(
			tc.LeftUntilDone,
			prometheus.GaugeValue,
			float64(t.LeftUntilDone),
			id, t.Name, tracker,
		)

		tstats := make(map[string]transmission.TrackerStat)

		for _, tracker := range t.TrackerStats {
			if tr, exists := tstats[tracker.Host]; exists {
				tr.DownloadCount += tracker.DownloadCount
				tstats[tracker.Host] = tr
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
