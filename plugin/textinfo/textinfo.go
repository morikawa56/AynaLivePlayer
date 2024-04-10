package textinfo

import (
	"AynaLivePlayer/core/adapter"
	"AynaLivePlayer/core/events"
	"AynaLivePlayer/core/model"
	"AynaLivePlayer/gui"
	"AynaLivePlayer/gui/component"
	"AynaLivePlayer/pkg/config"
	"AynaLivePlayer/pkg/event"
	"AynaLivePlayer/pkg/i18n"
	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"
	"github.com/go-resty/resty/v2"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

const MODULE_PLUGIN_TEXTINFO = "plugin.textinfo"

const Template_Path = "./template/"
const Out_Path = "./txtinfo/"

type Template struct {
	Name string
	Text string
	Tmpl *template.Template
}

type MediaInfo struct {
	Index    int
	Title    string
	Artist   string
	Album    string
	Username string
	Cover    model.Picture
}

type OutInfo struct {
	Current       MediaInfo
	CurrentTime   int
	TotalTime     int
	Lyric         string
	Playlist      []MediaInfo
	PlaylistCount int
}

type TextInfo struct {
	config.BaseConfig
	Rendering  bool
	info       OutInfo
	templates  []*Template
	emptyCover []byte
	panel      fyne.CanvasObject
	ctr        adapter.IControlBridge
	log        adapter.ILogger
}

func NewTextInfo(ctr adapter.IControlBridge) *TextInfo {
	b, _ := ioutil.ReadFile(config.GetAssetPath("empty.png"))
	return &TextInfo{Rendering: true, emptyCover: b, ctr: ctr, log: ctr.Logger().WithModule(MODULE_PLUGIN_TEXTINFO)}
}

func (t *TextInfo) Title() string {
	return i18n.T("plugin.textinfo.title")
}

func (t *TextInfo) Description() string {
	return i18n.T("plugin.textinfo.description")
}

func (t *TextInfo) CreatePanel() fyne.CanvasObject {
	if t.panel != nil {
		return t.panel
	}
	enableRendering := container.NewHBox(
		widget.NewLabel(i18n.T("plugin.textinfo.prompt")),
		component.NewCheckOneWayBinding(
			i18n.T("plugin.textinfo.checkbox"),
			&t.Rendering, t.Rendering),
	)
	t.panel = container.NewVBox(enableRendering)
	return t.panel
}

func (t *TextInfo) Name() string {
	return "TextInfo"
}

func (t *TextInfo) Enable() (err error) {
	// ensure the output/input directory exists
	if err = os.MkdirAll(Template_Path, 0755); err != nil {
		return
	}
	if err = os.MkdirAll(Out_Path, 0755); err != nil {
		return
	}
	config.LoadConfig(t)
	t.reloadTemplates()
	t.registerHandlers()
	gui.AddConfigLayout(t)
	return nil
}

func (d *TextInfo) Disable() error {
	return nil
}

func (t *TextInfo) reloadTemplates() {
	var err error
	t.templates = make([]*Template, 0)
	files, err := ioutil.ReadDir(Template_Path)
	if err != nil {
		t.log.Warn("read template directory failed: ", err)
		return
	}
	for _, f := range files {
		t.log.Info("loading template: ", f.Name())
		content, err := ioutil.ReadFile(filepath.Join(Template_Path, f.Name()))
		if err != nil {
			t.log.Warnf("read template file %s failed: %s", f.Name(), err)
			continue
		}
		parse, err := template.New("info").
			Funcs(template.FuncMap{
				"GetSeconds": func(t int) int {
					return t % 60
				},
				"GetMinutes": func(t int) int {
					return t / 60
				},
			}).
			Parse(string(content))
		if err != nil {
			t.log.Warnf("parse template %s failed: %s", f.Name, err)
			continue
		}
		t.templates = append(t.templates, &Template{
			Name: f.Name(),
			Text: string(content),
			Tmpl: parse,
		})
	}
}

// RenderTemplates render the template to the output file
func (t *TextInfo) RenderTemplates() {
	if !t.Rendering {
		return
	}
	for _, tmpl := range t.templates {
		t.log.Debug("rendering template: ", tmpl.Name)
		out, err := os.Create(filepath.Join(Out_Path, tmpl.Name))
		defer out.Close()
		if err != nil {
			t.log.Warnf("create output file %s failed: %s", tmpl.Name, err)
			continue
		}
		if err = tmpl.Tmpl.Execute(out, t.info); err != nil {
			t.log.Warnf("rendering template %s failed: %s", tmpl.Name, err)
			return
		}
	}
}

func (t *TextInfo) OutputCover() {
	if !t.Rendering {
		return
	}
	if !t.info.Current.Cover.Exists() {
		err := ioutil.WriteFile(filepath.Join(Out_Path, "cover.jpg"), t.emptyCover, 0666)
		if err != nil {
			t.log.Warnf("write cover file failed: %s", err)
		}
		return
	}
	if t.info.Current.Cover.Data != nil {
		err := ioutil.WriteFile(filepath.Join(Out_Path, "cover.jpg"), t.info.Current.Cover.Data, 0666)
		if err != nil {
			t.log.Warnf("write cover file failed: %s", err)
		}
		return
	}
	go func() {
		resp, err := resty.New().R().
			Get(t.info.Current.Cover.Url)
		if err != nil {
			t.log.Warnf("get cover %s content failed: %s", t.info.Current.Cover.Url, err)
			return
		}
		err = ioutil.WriteFile(filepath.Join(Out_Path, "cover.jpg"), resp.Body(), 0666)
		if err != nil {
			t.log.Warnf("write cover file failed: %s", err)
		}
	}()
}

func (t *TextInfo) registerHandlers() {
	t.ctr.PlayControl().EventManager().RegisterA(events.EventPlay, "plugin.textinfo.current", func(event *event.Event) {
		t.info.Current = MediaInfo{
			Index:    0,
			Title:    event.Data.(events.PlayEvent).Media.Title,
			Artist:   event.Data.(events.PlayEvent).Media.Artist,
			Album:    event.Data.(events.PlayEvent).Media.Album,
			Cover:    event.Data.(events.PlayEvent).Media.Cover,
			Username: event.Data.(events.PlayEvent).Media.ToUser().Name,
		}
		t.RenderTemplates()
		t.OutputCover()
	})
	if t.ctr.PlayControl().GetPlayer().ObserveProperty(
		model.PlayerPropTimePos, "plugin.txtinfo.timepos", func(event *event.Event) {
			data := event.Data.(events.PlayerPropertyUpdateEvent).Value
			if data == nil {
				t.info.CurrentTime = 0
				return
			}
			ct := int(data.(float64))
			if ct == t.info.CurrentTime {
				return
			}
			t.info.CurrentTime = ct
			t.RenderTemplates()
		}) != nil {
		t.log.Error("register time-pos handler failed")
	}
	if t.ctr.PlayControl().GetPlayer().ObserveProperty(
		model.PlayerPropDuration, "plugin.txtinfo.duration", func(event *event.Event) {
			data := event.Data.(events.PlayerPropertyUpdateEvent).Value
			if data == nil {
				t.info.TotalTime = 0
				return
			}
			t.info.TotalTime = int(data.(float64))
			t.RenderTemplates()
		}) != nil {
		t.log.Error("fail to register handler for total time with property duration")
	}
	t.ctr.Playlists().GetCurrent().EventManager().RegisterA(
		events.EventPlaylistUpdate, "plugin.textinfo.playlist", func(event *event.Event) {
			pl := make([]MediaInfo, 0)
			e := event.Data.(events.PlaylistUpdateEvent)
			for index, m := range e.Playlist.Medias {
				pl = append(pl, MediaInfo{
					Index:    index,
					Title:    m.Title,
					Artist:   m.Artist,
					Album:    m.Album,
					Username: m.ToUser().Name,
				})
			}
			t.info.Playlist = pl
			t.RenderTemplates()
		})
	t.ctr.PlayControl().GetLyric().EventManager().RegisterA(
		events.EventLyricUpdate, "plugin.textinfo.lyric", func(event *event.Event) {
			lrcLine := event.Data.(events.LyricUpdateEvent).Lyric
			t.info.Lyric = lrcLine.Now.Lyric
			t.RenderTemplates()
		})

}
