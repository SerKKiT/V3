package transcoder

// Profile представляет конфигурацию одного качества видео
type Profile struct {
	Name         string
	Resolution   string
	Width        int
	Height       int
	VideoBitrate string
	MaxRate      string
	BufSize      string
	AudioBitrate string
	Framerate    int
}

// ABRConfig представляет конфигурацию Adaptive Bitrate Streaming
type ABRConfig struct {
	Profiles     []Profile
	SegmentTime  int    // Длительность сегмента в секундах
	PlaylistSize int    // Размер playlist (0 = все сегменты)
	PlaylistType string // "event" для live
}

// DefaultABRProfiles - набор качеств для адаптивного стриминга
var DefaultABRProfiles = []Profile{
	{
		Name:         "1080p",
		Resolution:   "1920x1080",
		Width:        1920,
		Height:       1080,
		VideoBitrate: "5000k",
		MaxRate:      "5500k",
		BufSize:      "11000k",
		AudioBitrate: "192k",
		Framerate:    30,
	},
	{
		Name:         "720p",
		Resolution:   "1280x720",
		Width:        1280,
		Height:       720,
		VideoBitrate: "2800k",
		MaxRate:      "3080k",
		BufSize:      "5600k",
		AudioBitrate: "128k",
		Framerate:    30,
	},
	{
		Name:         "480p",
		Resolution:   "854x480",
		Width:        854,
		Height:       480,
		VideoBitrate: "1400k",
		MaxRate:      "1540k",
		BufSize:      "2800k",
		AudioBitrate: "128k",
		Framerate:    30,
	},
	{
		Name:         "360p",
		Resolution:   "640x360",
		Width:        640,
		Height:       360,
		VideoBitrate: "800k",
		MaxRate:      "880k",
		BufSize:      "1600k",
		AudioBitrate: "96k",
		Framerate:    30,
	},
}

// DefaultABRConfig - конфигурация по умолчанию для live стриминга
var DefaultABRConfig = ABRConfig{
	Profiles:     DefaultABRProfiles,
	SegmentTime:  4,
	PlaylistSize: 0,
	PlaylistType: "event",
}

// GetProfileNames возвращает список имён профилей
func GetProfileNames() []string {
	names := make([]string, len(DefaultABRProfiles))
	for i, p := range DefaultABRProfiles {
		names[i] = p.Name
	}
	return names
}
