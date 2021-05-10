package naming

type Prefix interface {
	Args() []string
}

func build(prefix Prefix, needPrefix bool) string {
	if !needPrefix {
		return buildName(prefix.Args()[1:]...)
	}
	return buildName(prefix.Args()...)
}
func buildName(args ...string) (res string) {
	if len(args) == 0 {
		return ""
	}
	for i, arg := range args {
		if i == len(args)-1 {
			res += arg
		} else {
			res += ":" + arg
		}
	}
	return
}
