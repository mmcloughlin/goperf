package parse

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"

	"golang.org/x/perf/storage/benchfmt"
)

// Result is a benchmark result.
type Result struct {
	// FullName is the complete name of the benchmark, including parameters.
	FullName string
	// Name of the benchmark, excluding parameters and the "Benchmark" prefix.
	Name string
	// Parameters of the benchmark, extracted from the name.
	Parameters map[string]string
	// Labels are the persistent labels that applied to the result.
	Labels map[string]string
	// Value measured by the benchmark.
	Value float64
	// Unit of the measured value.
	Unit string
	// Line number on which the result was found.
	Line int
}

// Bytes parses results from b.
func Bytes(b []byte) ([]*Result, error) {
	return Reader(bytes.NewReader(b))
}

// Reader parses results from the supplied reader.
func Reader(r io.Reader) ([]*Result, error) {
	br := benchfmt.NewReader(r)
	var all []*Result
	for br.Next() {
		results, err := convert(br.Result())
		if err != nil {
			return nil, err
		}
		all = append(all, results...)
	}
	if err := br.Err(); err != nil {
		return nil, err
	}
	return all, nil
}

// convert parsed line into multiple results.
func convert(res *benchfmt.Result) ([]*Result, error) {
	line, err := parseresultline(res.Content)
	if err != nil {
		return nil, err
	}

	params := res.NameLabels.Copy()
	name := params["name"]
	delete(params, "name")

	var rs []*Result
	for _, m := range line.measurements {
		rs = append(rs, &Result{
			FullName:   line.name,
			Name:       name,
			Parameters: params,
			Labels:     res.Labels,
			Value:      m.value,
			Unit:       m.unit,
			Line:       res.LineNum,
		})
	}
	return rs, nil
}

// Reference: https://github.com/golang/proposal/blob/85effef2002473b4bb7d08f4adc3dd5b7449a82d/design/14313-benchmark-format.md#L119-L149
//
//	### Benchmark Results
//
//	A benchmark result line has the general form
//
//		<name> <iterations> <value> <unit> [<value> <unit>...]
//
//	The fields are separated by runs of space characters (as defined by `unicode.IsSpace`),
//	so the line can be parsed with `strings.Fields`.
//	The line must have an even number of fields, and at least four.
//
//	The first field is the benchmark name, which must begin with `Benchmark`
//	followed by an upper case character (as defined by `unicode.IsUpper`)
//	or the end of the field,
//	as in `BenchmarkReverseString` or just `Benchmark`.
//	Tools displaying benchmark data conventionally omit the `Benchmark` prefix.
//	The same benchmark name can appear on multiple result lines,
//	indicating that the benchmark was run multiple times.
//
//	The second field gives the number of iterations run.
//	For most processing this number can be ignored, although
//	it may give some indication of the expected accuracy
//	of the measurements that follow.
//
//	The remaining fields report value/unit pairs in which the value
//	is a float64 that can be parsed by `strconv.ParseFloat`
//	and the unit explains the value, as in “64.88 MB/s”.
//	The units reported are typically normalized so that they can be
//	interpreted without considering to the number of iterations.
//	In the example, the CPU cost is reported per-operation and the
//	throughput is reported per-second; neither is a total that
//	depends on the number of iterations.
//

type measurement struct {
	value float64
	unit  string
}

type resultline struct {
	name         string
	iterations   uint64
	measurements []measurement
}

func parseresultline(s string) (*resultline, error) {
	r := &resultline{}

	// Break into fields.
	fields := strings.Fields(s)
	if len(fields) < 4 {
		return nil, errors.New("result line must have at least four fields")
	}
	if len(fields)%2 != 0 {
		return nil, errors.New("result line must have an even number of fields")
	}

	// Name field.
	if !strings.HasPrefix(fields[0], "Benchmark") {
		return nil, errors.New("benchmark name must begin with \"Benchmark\"")
	}
	r.name = fields[0]

	// Iterations.
	iter, err := strconv.ParseUint(fields[1], 10, 64)
	if err != nil {
		return nil, fmt.Errorf("could not parse field %q as iterations", fields[1])
	}
	r.iterations = iter

	// Value/unit pairs.
	for i := 2; i < len(fields); i += 2 {
		value := fields[i]
		unit := fields[i+1]

		v, err := strconv.ParseFloat(value, 64)
		if err != nil {
			return nil, fmt.Errorf("could not parse field %q as float value", value)
		}

		r.measurements = append(r.measurements, measurement{
			value: v,
			unit:  unit,
		})
	}

	return r, nil
}
