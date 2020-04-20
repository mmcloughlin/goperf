package sched

// smoothstep computes a function that smoothly transitions from y0 at x0 to y1
// at x1.
func smoothstep(x, x0, y0, x1, y1 float64) float64 {
	return y0 + (y1-y0)*smoothunitstep(x, x0, x1)
}

// smoothunitstep computes a function that transitions from 0 at x0 to 1 at x1.
// Reference: https://en.wikipedia.org/wiki/Smoothstep.
func smoothunitstep(x, x0, x1 float64) float64 {
	x = clamp((x-x0)/(x1-x0), 0, 1)
	return x * x * x * (x*(x*6-15) + 10)
}

func clamp(x, lower, upper float64) float64 {
	if x < lower {
		return lower
	}
	if x > upper {
		return upper
	}
	return x
}
