package router

import (
	"sync"

	"github.com/pingostack/pingos/core/peer"
)

type Router struct {
	id         string
	publisher  *peer.Publisher
	subscriber []*peer.Subscriber
	lock       sync.RWMutex
}

func NewRouter(id string) *Router {
	return &Router{
		id: id,
	}
}

func (r *Router) GetID() string {
	return r.id
}

func (r *Router) GetPublisher() *peer.Publisher {
	return r.publisher
}

func (r *Router) GetSubscribers() []*peer.Subscriber {
	return r.subscriber
}
