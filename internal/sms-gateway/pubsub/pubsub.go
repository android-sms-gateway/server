package pubsub

import "github.com/go-core-fx/pubsubfx/pubsub"

// PubSub is the application's publish/subscribe interface: pubsubfx's PubSub,
// aliased so consumers depend on the application module.
type PubSub = pubsub.PubSub
