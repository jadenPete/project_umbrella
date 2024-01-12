package loader

import "project_umbrella/interpreter/runtime/value"

type LoaderChannel struct {
	LoadRequest  chan *LoaderRequest
	LoadResponse chan value.Value
}

func (channel *LoaderChannel) Close() {
	close(channel.LoadRequest)
	close(channel.LoadResponse)
}

func NewLoaderChannel() *LoaderChannel {
	return &LoaderChannel{
		LoadRequest:  make(chan *LoaderRequest),
		LoadResponse: make(chan value.Value),
	}
}

type LoaderRequest struct {
	Type LoaderRequestType
	Name string
}

type LoaderRequestType int

const (
	ModuleRequest LoaderRequestType = iota + 1
	LibraryRequest
)
