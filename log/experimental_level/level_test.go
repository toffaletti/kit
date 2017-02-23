package level_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/experimental_level"
)

func TestVariousLevels(t *testing.T) {
	for _, testcase := range []struct {
		allowed level.Option
		want    string
	}{
		{
			level.AllowAll(),
			strings.Join([]string{
				`{"level":"debug","this is":"debug log"}`,
				`{"level":"info","this is":"info log"}`,
				`{"level":"warn","this is":"warn log"}`,
				`{"level":"error","this is":"error log"}`,
			}, "\n"),
		},
		{
			level.AllowDebugAndAbove(),
			strings.Join([]string{
				`{"level":"debug","this is":"debug log"}`,
				`{"level":"info","this is":"info log"}`,
				`{"level":"warn","this is":"warn log"}`,
				`{"level":"error","this is":"error log"}`,
			}, "\n"),
		},
		{
			level.AllowInfoAndAbove(),
			strings.Join([]string{
				`{"level":"info","this is":"info log"}`,
				`{"level":"warn","this is":"warn log"}`,
				`{"level":"error","this is":"error log"}`,
			}, "\n"),
		},
		{
			level.AllowWarnAndAbove(),
			strings.Join([]string{
				`{"level":"warn","this is":"warn log"}`,
				`{"level":"error","this is":"error log"}`,
			}, "\n"),
		},
		{
			level.AllowErrorOnly(),
			strings.Join([]string{
				`{"level":"error","this is":"error log"}`,
			}, "\n"),
		},
		{
			level.AllowNone(),
			``,
		},
	} {
		var buf bytes.Buffer
		logger := level.NewFilter(log.NewJSONLogger(&buf), testcase.allowed)

		level.Debug(logger).Log("this is", "debug log")
		level.Info(logger).Log("this is", "info log")
		level.Warn(logger).Log("this is", "warn log")
		level.Error(logger).Log("this is", "error log")

		if want, have := testcase.want, strings.TrimSpace(buf.String()); want != have {
			t.Errorf("given Allowed=%v: want\n%s\nhave\n%s", testcase.allowed, want, have)
		}
	}
}

func TestErrNotAllowed(t *testing.T) {
	myError := errors.New("squelched!")
	opts := []level.Option{
		level.AllowWarnAndAbove(),
		level.ErrNotAllowed(myError),
	}
	logger := level.NewFilter(log.NewNopLogger(), opts...)

	if want, have := myError, level.Info(logger).Log("foo", "bar"); want != have {
		t.Errorf("want %#+v, have %#+v", want, have)
	}

	if want, have := error(nil), level.Warn(logger).Log("foo", "bar"); want != have {
		t.Errorf("want %#+v, have %#+v", want, have)
	}
}

func TestErrNoLevel(t *testing.T) {
	myError := errors.New("no level specified")

	var buf bytes.Buffer
	opts := []level.Option{
		level.SquelchNoLevel(true),
		level.ErrNoLevel(myError),
	}
	logger := level.NewFilter(log.NewJSONLogger(&buf), opts...)

	if want, have := myError, logger.Log("foo", "bar"); want != have {
		t.Errorf("want %v, have %v", want, have)
	}
	if want, have := ``, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestAllowNoLevel(t *testing.T) {
	var buf bytes.Buffer
	opts := []level.Option{
		level.SquelchNoLevel(false),
		level.ErrNoLevel(errors.New("I should never be returned!")),
	}
	logger := level.NewFilter(log.NewJSONLogger(&buf), opts...)

	if want, have := error(nil), logger.Log("foo", "bar"); want != have {
		t.Errorf("want %v, have %v", want, have)
	}
	if want, have := `{"foo":"bar"}`, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestLevelContext(t *testing.T) {
	var buf bytes.Buffer

	// Wrapping the level logger with a context allows users to use
	// log.DefaultCaller as per normal.
	var logger log.Logger
	logger = log.NewLogfmtLogger(&buf)
	logger = level.NewFilter(logger, level.AllowAll())
	logger = log.NewContext(logger).With("caller", log.DefaultCaller)

	level.Info(logger).Log("foo", "bar")
	if want, have := `level=info caller=level_test.go:137 foo=bar`, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}

func TestContextLevel(t *testing.T) {
	var buf bytes.Buffer

	// Wrapping a context with the level logger still works, but requires users
	// to specify a higher callstack depth value.
	var logger log.Logger
	logger = log.NewLogfmtLogger(&buf)
	logger = log.NewContext(logger).With("caller", log.Caller(5))
	logger = level.NewFilter(logger, level.AllowAll())

	level.Info(logger).Log("foo", "bar")
	if want, have := `caller=level_test.go:153 level=info foo=bar`, strings.TrimSpace(buf.String()); want != have {
		t.Errorf("want %q, have %q", want, have)
	}
}
