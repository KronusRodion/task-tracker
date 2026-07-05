package ctxkeys

type contextKey string

const (
	UserKey  contextKey = "user"
	ExecKey contextKey = "exec"
)
