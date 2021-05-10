package naming

type Exchange struct {
	prefix  string
	appName string
	kind    ExchangeKind
}

type ExchangeKind string

const (
	Direct ExchangeKind = "direct"
	Topic  ExchangeKind = "topic"
	Fanout ExchangeKind = "fanout"
	Header ExchangeKind = "header"
)

func NewExchange(appName string, kind ExchangeKind) *Exchange {
	return &Exchange{
		prefix:  "exchange",
		appName: appName,
		kind:    kind,
	}
}

func (e *Exchange) Args() []string {
	return []string{e.prefix, e.appName, e.kind.String()}
}

func (e *Exchange) Kind() string {
	return e.kind.String()
}

func (e *Exchange) String(prefix bool) string {
	return build(e, prefix)
}

func (e ExchangeKind) String() string {
	return string(e)
}
