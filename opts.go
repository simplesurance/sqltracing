package sqltracing

// Opt is a type for options for the Interceptor.
type Opt func(*Interceptor)

// WithOpsExcluded can be passed when creating an Interceptor.
// It excludes recording traces for the passed database operations.
func WithOpsExcluded(ops ...SQLOp) Opt {
	return func(drv *Interceptor) {
		for _, op := range ops {
			drv.excludedOps[op] = struct{}{}
		}
	}
}
