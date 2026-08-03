# Assert - A minimal Go test package

Inspired by:
* https://github.com/nalgeon/be
* https://antonz.org/do-not-testify/

## Usage

```go
package foo

import (
	"testing"

	"github.com/cactus/go-camo/v2/pkg/assert"
)

// errType is a custom error type.
type errType string

func (e errType) Error() string {
	return string(e)
}

func TestSomething(t *testing.T) {
    t.Parallel()

    err := errors.New("my error")
    assert.Nil(t, err) // test assertion that something is nil
    // output => &errors.errorString{s:"my error"}; want: <nil>;

    err = errType("oops")
    assert.NotNil(t, nil) // test assertion that something is NOT nil
    // output => got: <nil>; expected non-nil

    // assert that an error value matches (string match)
    assert.ErrorIs(t, err, "my bad error")
    // output => got: "oops"; want: "my bad error";

    assert.ErrorIs(t, nil, err)  // assert error value matches (error match)
    // output => got: <nil>; want: errType(oops);

    wrappedErr := fmt.Errorf("wrapped: %w", err)
    assert.ErrorIs(t, nil, wrappedErr) // works with wrapped errors, using errors.Is under the hood
    // output => got: <nil>; want: *fmt.wrapError(wrapped: oops)

    // can also check for error type, using errors.As under the hood
    assert.ErrorIs(t, nil, reflect.TypeFor[*fs.PathError]())
    // output => got: <nil>; want: *fs.PathError

    // assert boolean true
    assert.True(t, false)
    // output => got: false; want: true;

    // assert boolean false
    assert.False(t, true)
    // output => got: true; want: false;

    // assert equal
    assert.Equal(t, 1, 2)
    // output => got: 1; want: 2;

    // assert NOT equal
    assert.NotEqual(t, 1, 1)
    // output => got: 1; expected values to be different;

    // assert string matches regex
    assert.MatchesRegexp(t, "abc123d", `abc[123]+$`)
    // output => got: "abc123d"; want to match "abc[123]+$";

    // a third argument can be passed to each of these functions to emit additional
    // information on failure
    assert.Equal(t, 1, 2,
        fmt.Sprintf("%d times around the moon", 3),
    )
    // output => got: 1; want: 2; 3 times around the moon

}
```
