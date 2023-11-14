package pts

// ChannelRegisterFunc is a function that is called to register a channel
type ChannelRegisterFunc func(channelName string, handlers ChannelHandlers) *Channel

// PluginInitializer is a function that is called to initialize a plugin.
type PluginInitializer func(registerChannel ChannelRegisterFunc) error

// Plugin contains all handler functions for a Plugin.
type Plugin struct {
	Init PluginInitializer
}
