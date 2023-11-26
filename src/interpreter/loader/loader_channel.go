package loader

import "project_umbrella/interpreter/runtime/value"

type LoaderChannel struct {
	LoadRequest  chan string
	LoadResponse chan value.Value
}

func (channel *LoaderChannel) Close() {
	close(channel.LoadRequest)
	close(channel.LoadResponse)
}

func NewLoaderChannel() *LoaderChannel {
	return &LoaderChannel{
		LoadRequest:  make(chan string),
		LoadResponse: make(chan value.Value),
	}
}
