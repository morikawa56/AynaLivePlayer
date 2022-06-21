package player

import "AynaLivePlayer/liveclient"

type Media struct {
	Title  string
	Artist string
	Cover  string
	Album  string
	Lyric  string
	Url    string
	Header map[string]string
	User   interface{}
	Meta   interface{}
}

func (m *Media) ToUser() *User {
	if u, ok := m.User.(*User); ok {
		return u
	}
	return &User{Name: m.DanmuUser().Username}
}

func (m *Media) SystemUser() *User {
	if u, ok := m.User.(*User); ok {
		return u
	}
	return nil
}

func (m *Media) DanmuUser() *liveclient.DanmuUser {
	if u, ok := m.User.(*liveclient.DanmuUser); ok {
		return u
	}
	return nil
}
