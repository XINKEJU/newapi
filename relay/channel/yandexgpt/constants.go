package yandexgpt

// ModelList contains officially supported YandexGPT model IDs.
// Users pass the short alias; the adaptor maps it to the full URI at request time.
var ModelList = []string{
	"yandexgpt",
	"yandexgpt-lite",
	"yandexgpt-32k",
	"yandexgpt-lite-rc",
	"yandexgpt-rc",
	"summarization",
}

var ChannelName = "yandexgpt"
