package platforms

// PlatformHandler is the messaging interface for platform bots.
type PlatformHandler interface {
	SendMessage(channelID, message string) error
	ChannelType() string
	GetUserInfo(userID string) (name, displayName string, err error)
	ResolveChannelID(targetID string) (string, error)
}

// Platform is a generic interface exposing the underlying SDK client.
type Platform[C any] interface {
	PlatformHandler
	RawClient() C
}
