package main

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/theme"
)

// customTheme wraps the default Fyne theme to apply an accent color and a
// dark/light variant without redefining every asset.
type customTheme struct {
	base   fyne.Theme
	accent color.Color
	dark   bool
}

func newCustomTheme(accent string, dark bool) *customTheme {
	return &customTheme{
		base:   theme.DefaultTheme(),
		accent: parseColor(accent),
		dark:   dark,
	}
}

func (c *customTheme) Color(name fyne.ThemeColorName, _ fyne.ThemeVariant) color.Color {
	switch name {
	case theme.ColorNamePrimary, theme.ColorNameHyperlink, theme.ColorNameSelection:
		return c.accent
	case theme.ColorNameBackground:
		if c.dark {
			return color.NRGBA{R: 0x12, G: 0x14, B: 0x18, A: 0xff}
		}
		return color.NRGBA{R: 0xf7, G: 0xf8, B: 0xfa, A: 0xff}
	case theme.ColorNameForeground:
		if c.dark {
			return color.NRGBA{R: 0xe6, G: 0xe9, B: 0xef, A: 0xff}
		}
		return color.NRGBA{R: 0x1f, G: 0x24, B: 0x2e, A: 0xff}
	case theme.ColorNameButton:
		if c.dark {
			return color.NRGBA{R: 0x1f, G: 0x24, B: 0x2c, A: 0xff}
		}
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameInputBackground:
		if c.dark {
			return color.NRGBA{R: 0x1a, G: 0x1d, B: 0x24, A: 0xff}
		}
		return color.NRGBA{R: 0xff, G: 0xff, B: 0xff, A: 0xff}
	case theme.ColorNameHover:
		if c.dark {
			return color.NRGBA{R: 0x2a, G: 0x30, B: 0x3a, A: 0xff}
		}
		return color.NRGBA{R: 0xe9, G: 0xec, B: 0xf1, A: 0xff}
	case theme.ColorNameSeparator:
		if c.dark {
			return color.NRGBA{R: 0x2c, G: 0x32, B: 0x3c, A: 0xff}
		}
		return color.NRGBA{R: 0xd9, G: 0xde, B: 0xe6, A: 0xff}
	}
	return c.base.Color(name, variantFor(c.dark))
}

func (c *customTheme) Icon(name fyne.ThemeIconName) fyne.Resource {
	return c.base.Icon(name)
}

func (c *customTheme) Font(style fyne.TextStyle) fyne.Resource {
	return c.base.Font(style)
}

func (c *customTheme) Size(name fyne.ThemeSizeName) float32 {
	switch name {
	case theme.SizeNamePadding:
		return 6
	case theme.SizeNameInnerPadding:
		return 8
	case theme.SizeNameText:
		return 14
	case theme.SizeNameHeadingText:
		return 18
	}
	return c.base.Size(name)
}

func variantFor(dark bool) fyne.ThemeVariant {
	if dark {
		return theme.VariantDark
	}
	return theme.VariantLight
}

func parseColor(hex string) color.Color {
	if len(hex) == 0 {
		return color.NRGBA{R: 0x25, G: 0x63, B: 0xeb, A: 0xff}
	}
	if hex[0] == '#' {
		hex = hex[1:]
	}
	var r, g, b uint8
	if len(hex) == 6 {
		fmtSscan(hex[0:2], &r)
		fmtSscan(hex[2:4], &g)
		fmtSscan(hex[4:6], &b)
	}
	return color.NRGBA{R: r, G: g, B: b, A: 0xff}
}

// fmtSscan parses a two-char hex byte without pulling in fmt/strconv at every
// color lookup; kept tiny and explicit.
func fmtSscan(s string, v *uint8) {
	*v = hexByte(s)
}

func hexByte(s string) uint8 {
	return uint8(hexVal(s[0])<<4 | hexVal(s[1]))
}

func hexVal(c byte) uint8 {
	switch {
	case c >= '0' && c <= '9':
		return c - '0'
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10
	}
	return 0
}
