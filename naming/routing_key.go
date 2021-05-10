package naming

type RoutingKey struct {
	prefix   string
	appName  string
	consumer Consumer
}

func NewRoutingKey(appName string, consumer Consumer) *RoutingKey {
	return &RoutingKey{
		prefix:   "key",
		appName:  appName,
		consumer: consumer,
	}
}

func (r *RoutingKey) Args() []string {
	return []string{r.prefix, r.appName, r.consumer.String()}
}

func (r *RoutingKey) Consumer() Consumer {
	return r.consumer
}

func (r *RoutingKey) String(prefix bool) string {
	return build(r, prefix)
}
