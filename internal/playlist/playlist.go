package playlist

import (
	"AynaLivePlayer/core/events"
	"AynaLivePlayer/core/model"
	"AynaLivePlayer/global"
	"AynaLivePlayer/pkg/config"
	"AynaLivePlayer/pkg/event"
	"AynaLivePlayer/pkg/logger"
	"github.com/AynaLivePlayer/miaosic"
)

var PlayerPlaylist *playlist = nil
var HistoryPlaylist *playlist = nil
var SystemPlaylist *playlist = nil
var PlaylistsPlaylist *playlist = nil

type playlistConfig struct {
	SystemPlaylistMode model.PlaylistMode
	SystemPlaylistID   string
	PlaylistsPath      string
	playlists          []miaosic.Playlist
}

func (p *playlistConfig) Name() string {
	return "Playlist"
}

func (p *playlistConfig) OnLoad() {
	err := config.LoadJson(p.PlaylistsPath, &p.playlists)
	if err != nil {
		log.Errorf("Failed to load playlists: %s", err.Error())
	}
	log.Infof("Loaded %d playlists", len(p.playlists))
}

func (p *playlistConfig) OnSave() {
	_ = config.SaveJson(p.PlaylistsPath, p.playlists)
	return
}

var cfg = &playlistConfig{
	SystemPlaylistMode: model.PlaylistModeNormal,
	PlaylistsPath:      "playlists.json",
	playlists:          make([]miaosic.Playlist, 0),
}

var log logger.ILogger = nil

func Initialize() {
	log = global.Logger.WithPrefix("Playlists")
	PlayerPlaylist = newPlaylist(model.PlaylistIDPlayer)
	SystemPlaylist = newPlaylist(model.PlaylistIDSystem)
	HistoryPlaylist = newPlaylist(model.PlaylistIDHistory)
	config.LoadConfig(cfg)

	global.EventManager.RegisterA(events.PlaylistModeChangeCmd(model.PlaylistIDSystem), "internal.playlist.system_init", func(event *event.Event) {
		cfg.SystemPlaylistMode = event.Data.(events.PlaylistModeChangeUpdateEvent).Mode
	})

	global.EventManager.CallA(events.PlaylistModeChangeUpdate(model.PlaylistIDSystem), events.PlaylistModeChangeUpdateEvent{
		Mode: cfg.SystemPlaylistMode,
	})

	global.EventManager.RegisterA(
		events.PlayerPlayingUpdate,
		"internal.playlist.player_playing_update",
		func(event *event.Event) {
			if event.Data.(events.PlayerPlayingUpdateEvent).Removed {
				return
			}
			global.EventManager.CallA(events.PlaylistInsertCmd(model.PlaylistIDHistory), events.PlaylistInsertCmdEvent{
				Media:    event.Data.(events.PlayerPlayingUpdateEvent).Media,
				Position: -1,
			})
		})

	createPlaylistManager()
}

func Close() {
	cfg.playlists = make([]miaosic.Playlist, 0)
	for _, v := range allPlaylists {
		cfg.playlists = append(cfg.playlists, *v)
	}
}
