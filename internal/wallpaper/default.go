package wallpaper

// DefaultSources returns the list of wallpaper sources tried when no image path
// is given. Order matters: first successful source wins.
func DefaultSources() []Source {
	return []Source{
		EnvSource(),
		Hyprland(),
		SwayBG(),
		Feh(),
		Nitrogen(),
		GNOME(),
	}
}

// CurrentWallpaperPath returns the current wallpaper path by trying default sources.
func CurrentWallpaperPath() (string, error) {
	path, _, err := FirstOf(DefaultSources()...)
	return path, err
}
