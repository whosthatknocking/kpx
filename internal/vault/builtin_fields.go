package vault

import (
	"github.com/tobischo/gokeepasslib/v3"
	"github.com/tobischo/gokeepasslib/v3/wrappers"
)

const (
	fieldTitle    = "Title"
	fieldUserName = "UserName"
	fieldPassword = "Password"
	fieldURL      = "URL"
	fieldNotes    = "Notes"
)

func value(key, content string, protected bool) gokeepasslib.ValueData {
	return gokeepasslib.ValueData{
		Key: key,
		Value: gokeepasslib.V{
			Content:   content,
			Protected: wrappers.NewBoolWrapper(protected),
		},
	}
}

func setValue(entry *gokeepasslib.Entry, key, content string, protected bool) {
	index := entry.GetIndex(key)
	if index >= 0 {
		entry.Values[index].Value.Content = content
		entry.Values[index].Value.Protected = wrappers.NewBoolWrapper(protected)
		return
	}
	entry.Values = append(entry.Values, value(key, content, protected))
}

func removeValue(entry *gokeepasslib.Entry, key string) {
	index := entry.GetIndex(key)
	if index < 0 {
		return
	}
	entry.Values = append(entry.Values[:index], entry.Values[index+1:]...)
}

func isBuiltinField(key string) bool {
	switch key {
	case fieldTitle, fieldUserName, fieldPassword, fieldURL, fieldNotes:
		return true
	default:
		return false
	}
}
