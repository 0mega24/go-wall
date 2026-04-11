package wallpaper

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeSource struct {
	name string
	path string
	err  error
}

func (f fakeSource) Name() string { return f.name }
func (f fakeSource) WallpaperPath() (string, error) {
	if f.err != nil {
		return "", f.err
	}
	return f.path, nil
}

func TestFirstOf_FirstSucceeds(t *testing.T) {
	s1 := fakeSource{name: "a", err: errors.New("nope")}
	s2 := fakeSource{name: "b", path: "/tmp/wall.png"}
	s3 := fakeSource{name: "c", path: "/other.png"}
	path, name, err := FirstOf(s1, s2, s3)
	require.NoError(t, err)
	assert.Equal(t, "/tmp/wall.png", path)
	assert.Equal(t, "b", name)
}

func TestFirstOf_EmptyPathSkipped(t *testing.T) {
	s1 := fakeSource{name: "a", path: ""}
	s2 := fakeSource{name: "b", path: "/wall.png"}
	path, name, err := FirstOf(s1, s2)
	require.NoError(t, err)
	assert.Equal(t, "/wall.png", path)
	assert.Equal(t, "b", name)
}

func TestFirstOf_AllFail(t *testing.T) {
	s1 := fakeSource{name: "a", err: errors.New("e1")}
	s2 := fakeSource{name: "b", err: errors.New("e2")}
	_, _, err := FirstOf(s1, s2)
	assert.Error(t, err, "expected error when all sources fail")
}
