package plugin

type ActivePlugin interface {
	Name() string
	ID() string

	OnWebSocketRequest(data []byte)
	OnRequest(requestID string, data []byte)
}
