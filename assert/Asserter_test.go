package assert_test

import (
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/adamluzsi/testcase"
	"github.com/adamluzsi/testcase/assert"
	"github.com/adamluzsi/testcase/fixtures"
)

var _ testcase.Asserter = assert.Asserter{}

func asserter(failFn func(args ...interface{})) assert.Asserter {
	return assert.Asserter{FailFn: failFn, Helper: func() {}}
}

func Equal(tb testing.TB, a, b interface{}) {
	tb.Helper()
	if !reflect.DeepEqual(a, b) {
		tb.Fatalf("A and B not equal: %#v <=> %#v", a, b)
	}
}

func AssertFailFnArgs(tb testing.TB, expected, output []interface{}) {
	tb.Helper()
	tb.Logf("%#v %#v", output, expected)
	join := func(vs []interface{}) string {
		return strings.TrimSpace(fmt.Sprintln(vs...))
	}
	if strings.Contains(join(output), join(expected)) {
		return
	}
	tb.Fatalf("expected msg not found\noutput: %#v\nexpected: %#v", output, expected)
}

func TestAsserter_True(t *testing.T) {
	t.Run(`when true passed`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.True(true)
		Equal(t, failed, false)
	})
	t.Run(`when false passed`, func(t *testing.T) {
		var failed bool
		var actualMsg []interface{}
		subject := asserter(func(args ...interface{}) {
			failed = true
			actualMsg = args
		})
		expectedMsg := []interface{}{"foo", "bar", "baz"}
		subject.True(false, expectedMsg...)
		Equal(t, failed, true)
		AssertFailFnArgs(t, expectedMsg, actualMsg)
	})
}

func TestAsserter_Nil(t *testing.T) {
	t.Run(`when nil passed`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.Nil(nil)
		Equal(t, failed, false)
	})
	t.Run(`when non nil value is passed`, func(t *testing.T) {
		var failed bool
		var actualMsg []interface{}
		subject := asserter(func(args ...interface{}) {
			failed = true
			actualMsg = args
		})
		expectedMsg := []interface{}{"foo", "bar", "baz"}
		subject.Nil(errors.New("not nil"), expectedMsg...)
		Equal(t, failed, true)
		AssertFailFnArgs(t, expectedMsg, actualMsg)
	})
	t.Run("when non nil zero value is passed", func(t *testing.T) {
		var failed bool
		var actualMsg []interface{}
		subject := asserter(func(args ...interface{}) {
			failed = true
			actualMsg = args
		})
		expectedMsg := []interface{}{"foo", "bar", "baz"}
		subject.Nil("", expectedMsg...) // zero string value
		Equal(t, failed, true)
		AssertFailFnArgs(t, expectedMsg, actualMsg)
	})
}

func TestAsserter_NotNil(t *testing.T) {
	t.Run(`when nil passed`, func(t *testing.T) {
		var out []interface{}
		subject := asserter(func(args ...interface{}) { out = args })
		msg := []interface{}{"foo", "bar", "baz"}
		subject.NotNil(nil, msg...)
		AssertFailFnArgs(t, msg, out)
	})
	t.Run(`when non nil value is passed`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.NotNil(errors.New("not nil"), "foo", "bar", "baz")
		Equal(t, failed, false)
	})
	t.Run("when non nil zero value is passed", func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.NotNil("", "foo", "bar", "baz")
		Equal(t, failed, false)
	})
}

func TestAsserter_Panics(t *testing.T) {
	t.Run(`when no panic, fails`, func(t *testing.T) {
		var failed bool
		var out []interface{}
		subject := asserter(func(args ...interface{}) { failed = true; out = args })
		subject.Panic(func() { /* nothing */ }, "boom!")
		Equal(t, failed, true)
		AssertFailFnArgs(t, []interface{}{"boom!"}, out)
	})
	t.Run(`when panic with nil value, pass`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.Panic(func() { panic(nil) }, "boom!")
		Equal(t, failed, false)
	})
	t.Run(`when panic with something, pass`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.Panic(func() { panic("something") }, "boom!")
		Equal(t, failed, false)
	})
}

func TestAsserter_NotPanics(t *testing.T) {
	t.Run(`when no panic, pass`, func(t *testing.T) {
		var failed bool
		subject := asserter(func(args ...interface{}) { failed = true })
		subject.NotPanic(func() { /* nothing */ }, "boom!")
		Equal(t, failed, false)
	})
	t.Run(`when panic with nil value, fail`, func(t *testing.T) {
		var failed bool
		var out []interface{}
		subject := asserter(func(args ...interface{}) { failed = true; out = args })
		subject.NotPanic(func() { panic(nil) }, "boom!")
		Equal(t, failed, true)
		AssertFailFnArgs(t, []interface{}{"boom!"}, out)
	})
	t.Run(`when panic with something, fail`, func(t *testing.T) {
		var failed bool
		var out []interface{}
		subject := asserter(func(args ...interface{}) { failed = true; out = args })
		subject.NotPanic(func() { panic("something") }, "boom!")
		Equal(t, failed, true)
		AssertFailFnArgs(t, []interface{}{"boom!"}, out)
		AssertFailFnArgs(t, []interface{}{"something"}, out)
	})
}

func TestAsserter_Equal(t *testing.T) {
	type TestCase struct {
		Desc     string
		Expected interface{}
		Actual   interface{}
		IsFailed bool
	}
	type E struct{ V int }

	for _, tc := range []TestCase{
		{
			Desc:     "when two basic type provided - int - equals",
			Expected: 42,
			Actual:   42,
			IsFailed: false,
		},
		{
			Desc:     "when two basic type provided - int - not equal",
			Expected: 42,
			Actual:   24,
			IsFailed: true,
		},
		{
			Desc:     "when two basic type provided - string - equals",
			Expected: "42",
			Actual:   "42",
			IsFailed: false,
		},
		{
			Desc:     "when two basic type provided - string - not equal",
			Expected: "42",
			Actual:   "24",
			IsFailed: true,
		},
		{
			Desc:     "when struct is provided - equals",
			Expected: E{V: 42},
			Actual:   E{V: 42},
			IsFailed: false,
		},
		{
			Desc:     "when struct is provided - not equal",
			Expected: E{V: 42},
			Actual:   E{V: 24},
			IsFailed: true,
		},
		{
			Desc:     "when struct ptr is provided - equals",
			Expected: &E{V: 42},
			Actual:   &E{V: 42},
			IsFailed: false,
		},
		{
			Desc:     "when struct ptr is provided - not equal",
			Expected: &E{V: 42},
			Actual:   &E{V: 24},
			IsFailed: true,
		},
		{
			Desc:     "when byte slice is provided - equals",
			Expected: []byte("foo"),
			Actual:   []byte("foo"),
			IsFailed: false,
		},
		{
			Desc:     "when byte slice is provided - not equal",
			Expected: []byte("foo"),
			Actual:   []byte("bar"),
			IsFailed: true,
		},
		{
			Desc:     "when byte slice is provided - not equal - expected populated, actual nil",
			Expected: []byte("foo"),
			Actual:   nil,
			IsFailed: true,
		},
		{
			Desc:     "when byte slice is provided - not equal - expected nil, actual populated",
			Expected: nil,
			Actual:   []byte("foo"),
			IsFailed: true,
		},
	} {
		tc := tc
		t.Run(tc.Desc, func(t *testing.T) {
			expectedMsg := []interface{}{fixtures.Random.StringN(3), fixtures.Random.StringN(3)}
			var actualMsg []interface{}
			var failed bool
			subject := asserter(func(args ...interface{}) {
				failed = true
				actualMsg = args
			})

			subject.Equal(tc.Expected, tc.Actual, expectedMsg...)
			if tc.IsFailed && actualMsg != nil {
				t.Log(actualMsg...)
			}

			Equal(t, failed, tc.IsFailed)
			if !tc.IsFailed {
				return
			}

			AssertFailFnArgs(t, expectedMsg, actualMsg)
		})
	}
}

func TestAsserter_Equal_function(t *testing.T) {
	expectedMsg := []interface{}{fixtures.Random.StringN(3), fixtures.Random.StringN(3)}
	var actualMsg []interface{}
	var failed bool
	subject := asserter(func(args ...interface{}) {
		failed = true
		actualMsg = args
	})

	subject.Equal(func() {}, func() {}, expectedMsg...)
	Equal(t, true, failed)

	AssertFailFnArgs(t, []interface{}{"Value is expected to be equable"}, actualMsg)
}

func AssertContainsWith(tb testing.TB, isFailed bool, contains func(a assert.Asserter, msg []interface{})) {
	tb.Helper()

	expectedMsg := []interface{}{fixtures.Random.StringN(3), fixtures.Random.StringN(3)}
	var actualMsg []interface{}
	var failed bool
	subject := asserter(func(args ...interface{}) {
		failed = true
		actualMsg = args
	})

	contains(subject, expectedMsg)
	if isFailed && actualMsg != nil {
		tb.Log(actualMsg...)
	}

	Equal(tb, failed, isFailed)
	if !isFailed {
		return
	}

	// at this point slice contains behavior is confirmed
	AssertFailFnArgs(tb, expectedMsg, actualMsg)
}

func AssertContainsTestCase(src, has interface{}, isFailed bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		AssertContainsWith(t, isFailed, func(a assert.Asserter, msg []interface{}) {
			a.Contain(src, has, msg...)
		})
	}
}

func AssertContainExactlyTestCase(src, oth interface{}, isFailed bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		AssertContainsWith(t, isFailed, func(a assert.Asserter, msg []interface{}) {
			a.ContainExactly(src, oth, msg...)
		})
	}
}

func TestAsserter_Contains_invalid(t *testing.T) {
	t.Run(`when source is invalid`, func(t *testing.T) {
		var out []interface{}
		asserter(func(args ...interface{}) { out = args }).Contain(nil, []int{42})
		AssertFailFnArgs(t, []interface{}{"invalid source value"}, out)
	})
	t.Run(`when "has" is invalid`, func(t *testing.T) {
		var out []interface{}
		asserter(func(args ...interface{}) { out = args }).Contain([]int{42}, nil)
		AssertFailFnArgs(t, []interface{}{`invalid "has" value`}, out)
	})
}

func TestAsserter_Contains_typeMismatch(t *testing.T) {
	assert.Must(t).Panic(func() {
		asserter(func(args ...interface{}) {}).Contain([]int{42}, []string{"42"})
	}, "will panic on type mismatch")
	assert.Must(t).Panic(func() {
		asserter(func(args ...interface{}) {}).Contain([]int{42}, "42")
	}, "will panic on type mismatch")
}

func TestAsserter_Contains_sliceHasSubSlice(t *testing.T) {
	type TestCase struct {
		Desc     string
		Slice    interface{}
		Contains interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "int: when equals",
			Slice:    []int{42, 24},
			Contains: []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "int: when doesn't have all the elements",
			Slice:    []int{42, 24},
			Contains: []int{42, 24, 42},
			IsFailed: true,
		},
		{
			Desc:     "int: when fully includes in the beginning",
			Slice:    []int{42, 24, 4, 2, 2, 4},
			Contains: []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "int: when fully includes in the end",
			Slice:    []int{4, 2, 2, 4, 42, 24},
			Contains: []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "int: when fully includes in the middle",
			Slice:    []int{4, 2, 42, 24, 2, 4},
			Contains: []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     "string: when equals",
			Slice:    []string{"42", "24"},
			Contains: []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "string: when doesn't have all the elements",
			Slice:    []string{"42", "24"},
			Contains: []string{"42", "24", "42"},
			IsFailed: true,
		},
		{
			Desc:     "string: when fully includes in the beginning",
			Slice:    []string{"42", "24", "4", "2", "2", "4"},
			Contains: []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "string: when fully includes in the end",
			Slice:    []string{"4", "2", "2", "4", "42", "24"},
			Contains: []string{"42", "24"},
			IsFailed: false,
		},
		{
			Desc:     "string: when fully includes in the middle",
			Slice:    []string{"4", "2", "42", "24", "2", "4"},
			Contains: []string{"42", "24"},
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, AssertContainsTestCase(tc.Slice, tc.Contains, tc.IsFailed))
	}
}

func TestAsserter_Contains_map(t *testing.T) {
	type TestCase struct {
		Desc     string
		Map      interface{}
		Has      interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "when equals",
			Map:      map[int]int{42: 42, 24: 24},
			Has:      map[int]int{42: 42, 24: 24},
			IsFailed: false,
		},
		{
			Desc:     "when doesn't have all the elements",
			Map:      map[int]int{42: 42, 24: 24},
			Has:      map[int]int{42: 42, 24: 24, 12: 12},
			IsFailed: true,
		},
		{
			Desc:     "when map contains sub map",
			Map:      map[int]int{42: 42, 24: 24},
			Has:      map[int]int{42: 42},
			IsFailed: false,
		},
		{
			Desc:     "when map contains sub map keys but with different value",
			Map:      map[int]int{42: 24, 24: 42},
			Has:      map[int]int{42: 42},
			IsFailed: true,
		},
		{
			Desc:     "when map contains sub map keys, and values are nil",
			Map:      map[int]*int{42: nil, 24: nil},
			Has:      map[int]*int{42: nil},
			IsFailed: false,
		},
		{
			Desc:     "when map contains sub map keys, and the key is nil",
			Map:      map[*int]int{nil: 42},
			Has:      map[*int]int{nil: 42},
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, AssertContainsTestCase(tc.Map, tc.Has, tc.IsFailed))
	}
}

func TestAsserter_Contains_sliceHasElement(t *testing.T) {
	type TestCase struct {
		Desc     string
		Slice    interface{}
		Contains interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "int: when doesn't have the element",
			Slice:    []int{42, 24},
			Contains: 12,
			IsFailed: true,
		},
		{
			Desc:     "int: when has the value in the beginning",
			Slice:    []int{42, 24, 4, 2, 2, 4},
			Contains: 42,
			IsFailed: false,
		},
		{
			Desc:     "int: when has the value includes in the end",
			Slice:    []int{4, 2, 2, 4, 42, 24},
			Contains: 42,
			IsFailed: false,
		},
		{
			Desc:     "int: when has the value in the middle",
			Slice:    []int{4, 2, 42, 24, 2, 4},
			Contains: 42,
			IsFailed: false,
		},

		{
			Desc:     "string: when doesn't have the element",
			Slice:    []string{"42", "24"},
			Contains: "12",
			IsFailed: true,
		},
		{
			Desc:     "string: when has the value in the beginning",
			Slice:    []string{"42", "24", "4", "2", "2", "4"},
			Contains: "42",
			IsFailed: false,
		},
		{
			Desc:     "string: when has the value includes in the end",
			Slice:    []string{"4", "2", "2", "4", "42", "24"},
			Contains: "42",
			IsFailed: false,
		},
		{
			Desc:     "string: when has the value in the middle",
			Slice:    []string{"4", "2", "42", "24", "2", "4"},
			Contains: "42",
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, AssertContainsTestCase(tc.Slice, tc.Contains, tc.IsFailed))
	}
}

func TestAsserter_Contains_sliceOfInterface(t *testing.T) {
	t.Run(`when value implements the interface`, AssertContainsTestCase([]testing.TB{t}, t, false))

	t.Run(`when value doesn't implement the interface`, func(t *testing.T) {
		assert.Must(t).Panic(func() {
			AssertContainsTestCase([]testing.TB{t}, 42, true)(t)
		})
	})
}

func TestAsserter_Contains_stringHasSub(t *testing.T) {
	type TestCase struct {
		Desc     string
		String   interface{}
		Sub      interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "when doesn't have sub",
			String:   "Hello, world!",
			Sub:      "foo",
			IsFailed: true,
		},
		{
			Desc:     "when includes in the beginning",
			String:   "Hello, world!",
			Sub:      "Hello,",
			IsFailed: false,
		},
		{
			Desc:     "when includes in the middle",
			String:   "Hello, world!",
			Sub:      ", wor",
			IsFailed: false,
		},
		{
			Desc:     "when includes in the end",
			String:   "Hello, world!",
			Sub:      "world!",
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, AssertContainsTestCase(tc.String, tc.Sub, tc.IsFailed))
	}

	t.Run(`when value is a string based type`, func(t *testing.T) {
		type StringBasedType string

		t.Run(`and source has value`, AssertContainsTestCase(StringBasedType("foo/bar/baz"), StringBasedType("bar"), false))
		t.Run(`and source doesn't have value`, AssertContainsTestCase(StringBasedType("foo/bar/baz"), StringBasedType("oth"), true))
	})
}

func TestAsserter_ContainExactly_invalid(t *testing.T) {
	t.Run(`when source is invalid`, func(t *testing.T) {
		out := assert.Must(t).Panic(func() {
			asserter(func(args ...interface{}) {}).ContainExactly(nil, []int{42})
		})
		AssertFailFnArgs(t, []interface{}{"invalid expected value"}, []interface{}{out.(string)})
	})
	t.Run(`when "has" is invalid`, func(t *testing.T) {
		out := assert.Must(t).Panic(func() {
			asserter(func(args ...interface{}) {}).ContainExactly([]int{42}, nil)
		})
		AssertFailFnArgs(t, []interface{}{`invalid actual value`}, []interface{}{out.(string)})
	})

	assert.Must(t).Panic(func() {
		asserter(func(args ...interface{}) {}).ContainExactly([]int{42}, nil)
	})
}

func TestAsserter_ContainExactly_map(t *testing.T) {
	type TestCase struct {
		Desc     string
		Map      interface{}
		Has      interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     "when equals",
			Map:      map[int]int{42: 42, 24: 24},
			Has:      map[int]int{42: 42, 24: 24},
			IsFailed: false,
		},
		{
			Desc:     "when doesn't have all the elements",
			Map:      map[int]int{42: 42, 24: 24},
			Has:      map[int]int{42: 42, 24: 24, 12: 12},
			IsFailed: true,
		},
		{
			Desc:     "when map contains sub map keys but with different value",
			Map:      map[int]int{42: 24, 24: 42},
			Has:      map[int]int{42: 42, 24: 24},
			IsFailed: true,
		},
		{
			Desc:     "when map contains sub map keys, and values are nil",
			Map:      map[int]*int{42: nil, 24: nil},
			Has:      map[int]*int{42: nil, 24: nil},
			IsFailed: false,
		},
		{
			Desc:     "when map contains sub map keys, and the key is nil",
			Map:      map[*int]int{nil: 42},
			Has:      map[*int]int{nil: 42},
			IsFailed: false,
		},
	} {
		t.Run(tc.Desc, AssertContainExactlyTestCase(tc.Map, tc.Has, tc.IsFailed))
	}
}
func TestAsserter_ContainExactly_slice(t *testing.T) {
	type TestCase struct {
		Desc     string
		Src      interface{}
		Oth      interface{}
		IsFailed bool
	}

	for _, tc := range []TestCase{
		{
			Desc:     `when elements match with order`,
			Src:      []int{42, 24},
			Oth:      []int{42, 24},
			IsFailed: false,
		},
		{
			Desc:     `when elements match without order`,
			Src:      []int{42, 24},
			Oth:      []int{24, 42},
			IsFailed: false,
		},
		{
			Desc:     `when elements do not match`,
			Src:      []int{42, 24},
			Oth:      []int{4, 2, 2, 4},
			IsFailed: true,
		},
	} {
		t.Run(tc.Desc, AssertContainExactlyTestCase(tc.Src, tc.Oth, tc.IsFailed))
	}
}

func AssertNotContainTestCase(src, has interface{}, isFailed bool) func(*testing.T) {
	return func(t *testing.T) {
		t.Helper()

		AssertContainsWith(t, isFailed, func(a assert.Asserter, msg []interface{}) {
			a.NotContain(src, has, msg...)
		})
	}
}

func TestAsserter_NotContains(t *testing.T) {
	type TestCase struct {
		Desc        string
		Source      interface{}
		NotContains interface{}
		IsFailed    bool
	}

	for _, tc := range []TestCase{
		{
			Desc:        "when slice doesn't match elements",
			Source:      []int{42, 24},
			NotContains: []int{12},
			IsFailed:    false,
		},
		{
			Desc:        "when slice contain elements",
			Source:      []int{42, 24, 12},
			NotContains: []int{24, 12},
			IsFailed:    true,
		},
		{
			Desc:        "when map doesn't contains other map elements",
			Source:      map[int]int{42: 24},
			NotContains: map[int]int{12: 6},
			IsFailed:    false,
		},
		{
			Desc:        "when map contains other map elements",
			Source:      map[int]int{42: 24, 24: 12},
			NotContains: map[int]int{24: 12},
			IsFailed:    true,
		},
		{
			Desc:        "when map contains other map's key but with different value",
			Source:      map[int]int{42: 24, 24: 12},
			NotContains: map[int]int{24: 13},
			IsFailed:    false,
		},
		{
			Desc:        "when slice doesn't include the value",
			Source:      []int{42, 24},
			NotContains: 12,
			IsFailed:    false,
		},
		{
			Desc:        "when slice does include the value",
			Source:      []int{42, 24, 12},
			NotContains: 24,
			IsFailed:    true,
		},
	} {
		t.Run(tc.Desc, AssertNotContainTestCase(tc.Source, tc.NotContains, tc.IsFailed))
	}
}