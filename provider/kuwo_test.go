package provider

import (
	"AynaLivePlayer/player"
	"fmt"
	"testing"
)

func TestKuwo_Search(t *testing.T) {
	var api MediaProvider = KuwoAPI
	result, err := api.Search("染 reol")
	if err != nil {
		return
	}
	fmt.Println(result)
	media := result[0]
	err = api.UpdateMediaUrl(media)
	fmt.Println(err)
	fmt.Println(media.Url)
}

func TestKuwo_GetMusicMeta(t *testing.T) {
	var api MediaProvider = KuwoAPI

	media := player.Media{
		Meta: Meta{
			Name: api.GetName(),
			Id:   "22804772",
		},
	}
	err := api.UpdateMedia(&media)
	fmt.Println(err)
	if err != nil {
		return
	}
	fmt.Println(media)
}

func TestKuwo_GetMusic(t *testing.T) {
	var api MediaProvider = KuwoAPI
	media := player.Media{
		Meta: Meta{
			Name: api.GetName(),
			Id:   "22804772",
		},
	}
	err := api.UpdateMedia(&media)
	if err != nil {
		return
	}
	err = api.UpdateMediaUrl(&media)
	if err != nil {
		return
	}
	fmt.Println(media)
	fmt.Println(media.Url)
}

func TestKuwo_UpdateMediaLyric(t *testing.T) {
	var api MediaProvider = KuwoAPI
	media := player.Media{
		Meta: Meta{
			Name: api.GetName(),
			Id:   "22804772",
		},
	}
	err := api.UpdateMediaLyric(&media)
	fmt.Println(err)
	fmt.Println(media.Lyric)
}
