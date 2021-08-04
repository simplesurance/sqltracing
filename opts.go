package sqltracing

type Opt func(*Driver)

// WithOpsExcluded defines operations for which no traces are recorded.
func WithOpsExcluded(ops ...SQLOp) Opt {
	return func(drv *Driver) {
		for _, op := range ops {
			drv.excludedOps[op] = struct{}{}
		}
	}
}
