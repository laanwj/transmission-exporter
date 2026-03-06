package transmission

type (
	// TorrentCommand is the root command to interact with Transmission via RPC
	TorrentCommand struct {
		Method    string           `json:"method,omitempty"`
		Arguments TorrentArguments `json:"arguments,omitempty"`
		Result    string           `json:"result,omitempty"`
	}
	// TorrentArguments specifies the TorrentCommand in more detail
	TorrentArguments struct {
		Fields   []string  `json:"fields,omitempty"`
		Torrents []Torrent `json:"torrents,omitempty"`
		// Ids can also be an integer, but torrent IDs are not stable across Transmission restarts.
		// So we should only pass hashes or the special string "recently-active".
		// See https://github.com/transmission/transmission/blob/0fd35eb07032fb9a03dea23469f2d8e3abd43000/docs/rpc-spec.md#31-torrent-action-requests
		Ids          string                `json:"ids,omitempty"`
		DeleteData   bool                  `json:"delete-local-data,omitempty"`
		DownloadDir  string                `json:"download-dir,omitempty"`
		MetaInfo     string                `json:"metainfo,omitempty"`
		Filename     string                `json:"filename,omitempty"`
		TorrentAdded TorrentArgumentsAdded `json:"torrent-added"`
	}
	// TorrentArgumentsAdded specifies the torrent to get added data from
	TorrentArgumentsAdded struct {
		HashString string `json:"hashString"`
		ID         int64  `json:"id"`
		Name       string `json:"name"`
	}

	// Torrent represents a transmission torrent
	Torrent struct {
		ID                 int           `json:"id"`
		Name               string        `json:"name"`
		Status             int           `json:"status"`
		Added              int           `json:"addedDate"`
		ActivityDate       int           `json:"activityDate"`
		DoneDate           int           `json:"doneDate"`
		LeftUntilDone      int64         `json:"leftUntilDone"`
		Eta                int           `json:"eta"`
		UploadRatio        float64       `json:"uploadRatio"`
		RateDownload       int           `json:"rateDownload"`
		RateUpload         int           `json:"rateUpload"`
		DownloadDir        string        `json:"downloadDir"`
		IsFinished         bool          `json:"isFinished"`
		PercentDone        float64       `json:"percentDone"`
		SeedRatioMode      int           `json:"seedRatioMode"`
		HashString         string        `json:"hashString"`
		Error              int           `json:"error"`
		ErrorString        string        `json:"errorString"`
		Files              []File        `json:"files"`
		FilesStats         []FileStat    `json:"fileStats"`
		TrackerStats       []TrackerStat `json:"trackerStats"`
		Peers              []Peer        `json:"peers"`
		PeersConnected     int           `json:"peersConnected"`
		PeersGettingFromUs int           `json:"peersGettingFromUs"`
		PeersSendingToUs   int           `json:"peersSendingToUs"`
		TotalSize          int           `json:"totalSize"`
		SizeWhenDone       int64         `json:"sizeWhenDone"`
		DownloadEver       int           `json:"downloadedEver"`
		UploadedEver       int           `json:"uploadedEver"`
		CorruptEver        int64         `json:"corruptEver"`
		HaveValid          int64         `json:"haveValid"`
		DesiredAvailable   int64         `json:"desiredAvailable"`
		SecondsDownloading int64         `json:"secondsDownloading"`
		SecondsSeeding     int64         `json:"secondsSeeding"`
		QueuePosition      int           `json:"queuePosition"`
		PieceCount         int           `json:"pieceCount"`
		PieceSize          int           `json:"pieceSize"`
	}

	// File is a file contained inside a torrent
	File struct {
		BytesCompleted int64  `json:"bytesCompleted"`
		Length         int64  `json:"length"`
		Name           string `json:"name"`
	}

	// FileStat describe a file's priority & if it's wanted
	FileStat struct {
		BytesCompleted int64 `json:"bytesCompleted"`
		Priority       int   `json:"priority"`
		Wanted         bool  `json:"wanted"`
	}

	// TrackerStat has stats about the torrent's tracker
	TrackerStat struct {
		Announce              string `json:"announce"`
		AnnounceState         int    `json:"announceState"`
		DownloadCount         int    `json:"downloadCount"`
		HasAnnounced          bool   `json:"hasAnnounced"`
		HasScraped            bool   `json:"hasScraped"`
		Host                  string `json:"host"`
		ID                    int    `json:"id"`
		IsBackup              bool   `json:"isBackup"`
		LastAnnouncePeerCount int    `json:"lastAnnouncePeerCount"`
		LastAnnounceResult    string `json:"lastAnnounceResult"`
		LastAnnounceStartTime int    `json:"lastAnnounceStartTime"`
		LastAnnounceSucceeded bool   `json:"lastAnnounceSucceeded"`
		LastAnnounceTime      int    `json:"lastAnnounceTime"`
		LastAnnounceTimedOut  bool   `json:"lastAnnounceTimedOut"`
		LastScrapeResult      string `json:"lastScrapeResult"`
		LastScrapeStartTime   int    `json:"lastScrapeStartTime"`
		LastScrapeSucceeded   bool   `json:"lastScrapeSucceeded"`
		LastScrapeTime        int    `json:"lastScrapeTime"`
		LastScrapeTimedOut    bool   `json:"lastScrapeTimedOut"`
		LeecherCount          int    `json:"leecherCount"`
		NextAnnounceTime      int    `json:"nextAnnounceTime"`
		NextScrapeTime        int    `json:"nextScrapeTime"`
		Scrape                string `json:"scrape"`
		ScrapeState           int    `json:"scrapeState"`
		SeederCount           int    `json:"seederCount"`
		Tier                  int    `json:"tier"`
	}

	// Peer of a torrent
	Peer struct {
		Address            string  `json:"address"`
		ClientIsChoked     bool    `json:"clientIsChoked"`
		ClientIsInterested bool    `json:"clientIsInterested"`
		ClientName         string  `json:"clientName"`
		FlagStr            string  `json:"flagStr"`
		IsDownloadingFrom  bool    `json:"isDownloadingFrom"`
		IsEncrypted        bool    `json:"isEncrypted"`
		IsIncoming         bool    `json:"isIncoming"`
		IsUTP              bool    `json:"isUTP"`
		IsUploadingTo      bool    `json:"isUploadingTo"`
		PeerIsChoked       bool    `json:"peerIsChoked"`
		PeerIsInterested   bool    `json:"peerIsInterested"`
		Port               int     `json:"port"`
		Progress           float64 `json:"progress"`
		RateToClient       int     `json:"rateToClient"`
		RateToPeer         int     `json:"rateToPeer"`
	}
)

