package match

type Matcher interface {
	Match(error, interface{}) error
	Validate() error
}

var (
	_ Matcher = (*PayloadMatcher)(nil)
	_ Matcher = (*FnType)(nil)
)
