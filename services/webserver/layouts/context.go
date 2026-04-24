package layouts

import "context"

type contextKey string

const FragmentKey contextKey = "fragment"

func IsFragment(ctx context.Context) bool {
	v, ok := ctx.Value(FragmentKey).(bool)
	return ok && v
}
