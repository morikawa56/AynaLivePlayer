package internal

import (
	"AynaLivePlayer/internal/controller"
	"AynaLivePlayer/internal/liveroom"
	"AynaLivePlayer/internal/player"
	"AynaLivePlayer/internal/playlist"
	"AynaLivePlayer/internal/plugins"
	"AynaLivePlayer/internal/source"
	"AynaLivePlayer/plugin/diange"
	"AynaLivePlayer/plugin/qiege"
)

func Initialize() {
	player.SetupMpvPlayer()
	source.Initialize()
	playlist.Initialize()
	controller.Initialize()
	liveroom.Initialize()
	plugins.Initialize()
	plugins.LoadPlugins(diange.NewDiange(), qiege.NewQiege())
}

func Stop() {
	liveroom.StopAndSave()
	playlist.Close()
	player.StopMpvPlayer()
	plugins.ClosePlugins()
}
