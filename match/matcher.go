package match

type Matcher interface {
	Match(error, interface{}) error
	Validate() error
	// TODO add message type
}

var (
	_ Matcher = (*PayloadMatcher)(nil)
	_ Matcher = (*FnMatcher)(nil)
)
