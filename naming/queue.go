package naming

type Queue struct {
	prefix     string
	appName    string
	routingKey *RoutingKey
}

func NewQueue(appName string, routingKey *RoutingKey) *Queue {
	return &Queue{
		prefix:     "queue",
		appName:    appName,
		routingKey: routingKey,
	}
}
func (u *Queue) Args() []string {
	return []string{u.prefix, u.appName, u.routingKey.String(false)}
}

func (u *Queue) RoutingKey() *RoutingKey {
	return u.routingKey
}

func (u *Queue) String(prefix bool) string {
	return build(u, prefix)
}
