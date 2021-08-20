package sqltracing

type Opt func(*tracedDriver)

// WithOpsExcluded defines operations for which no traces are recorded.
func WithOpsExcluded(ops ...SQLOp) Opt {
	return func(drv *tracedDriver) {
		for _, op := range ops {
			drv.excludedOps[op] = struct{}{}
		}
	}
}
