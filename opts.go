package sqltracing

// Opt is a type for options that can be passed to NewDriver.
type Opt func(*tracedDriver)

// WithOpsExcluded can be passed as option to NewDriver.
// It excludes recording traces for the passed database operations.
func WithOpsExcluded(ops ...SQLOp) Opt {
	return func(drv *tracedDriver) {
		for _, op := range ops {
			drv.excludedOps[op] = struct{}{}
		}
	}
}
