package handlers

import (
	"fmt"
	"html/template"
)

type themeVars struct {
	BgBody       string
	BgSidebar    string
	BgContent    string
	BgCode       string
	BgCard       string
	BgCardHover  string
	TextPrimary  string
	TextSecondary string
	TextMuted    string
	BorderGlass  string
	BorderGlassHover string
	TableStripe  string
	InputBg      string
	InputBgFocus string
	HoverBg      string
	GlassWhite06 string // rgba(255,255,255,0.06) or dark equiv
	GlassWhite03 string
	GlassWhite04 string
	GlassWhite05 string
	GlassWhite10 string
	GlassWhite12 string
	HeadingStart string // gradient start for h1 headings (#ffffff on dark, dark on light)
	TextCode     string // text color inside code blocks and textareas
}

type accentVars struct {
	Accent1       string
	Accent2       string
	Accent3       string
	BtnGradEnd    string
	AccentDim     string
	GlowPurple    string
	GlowBlue      string
	HeadingTint   string // #a8c8ff equivalent
	BadgeBg       string // rgba(accent,0.12)
	BadgeBorder   string // rgba(accent,0.25)
	ActiveBg      string // rgba(accent,0.08)
	HoverBg       string // rgba(accent,0.06)
	FocusShadow   string // rgba(accent,0.15)
	CardGrad1     string // rgba(accent,0.15)
	CardGrad2     string // rgba(accent,0.1)
	CardBorder    string // rgba(accent,0.2)
	CardOverlay1  string // rgba(accent,0.08)
	CardOverlay2  string // rgba(accent2,0.04)
	BtnShadow     string // rgba(accent,0.4)
	AddCardHover  string // rgba(accent,0.06)
	AddCardShadow string // rgba(accent,0.1)
	IconGrad2Bg   string // rgba(accent2,0.15)
	IconGrad2Bdr  string // rgba(accent2,0.2)
	IconGrad3Bg   string // rgba(accent3,0.15)
	IconGrad3Bdr  string // rgba(accent3,0.2)
	IconGradMix   string // rgba(accent2,0.08) or rgba(accent,0.08)
	EditHoverBg   string // rgba(accent,0.12)
	VersionBdr    string // rgba(accent,0.2)
	IconBoxShadow string // rgba(accent,0.2)
	TableHeadBg   string // rgba(accent,0.12)
	TableHoverBg  string // rgba(accent,0.04)
	BlockquoteBg  string // rgba(accent,0.06)
	CopyHoverBdr  string // rgba(accent,0.3)
}

var themes = map[string]themeVars{
	"midnight": {
		BgBody: "#1a1d2e", BgSidebar: "#161929", BgContent: "#1e2236",
		BgCode: "#171a2a", BgCard: "rgba(255,255,255,0.06)", BgCardHover: "rgba(255,255,255,0.10)",
		TextPrimary: "#f0f0f5", TextSecondary: "#a3a9bc", TextMuted: "#6b7394",
		BorderGlass: "rgba(255,255,255,0.10)", BorderGlassHover: "rgba(255,255,255,0.20)",
		TableStripe: "rgba(255,255,255,0.03)", InputBg: "rgba(255,255,255,0.04)",
		InputBgFocus: "rgba(255,255,255,0.06)", HoverBg: "rgba(255,255,255,0.03)",
		GlassWhite06: "rgba(255,255,255,0.06)", GlassWhite03: "rgba(255,255,255,0.03)",
		GlassWhite04: "rgba(255,255,255,0.04)", GlassWhite05: "rgba(255,255,255,0.05)",
		GlassWhite10: "rgba(255,255,255,0.10)", GlassWhite12: "rgba(255,255,255,0.12)",
		HeadingStart: "#ffffff", TextCode: "#d6e4f0",
	},
	"slate": {
		BgBody: "#2d3148", BgSidebar: "#262a3e", BgContent: "#333750",
		BgCode: "#252840", BgCard: "rgba(255,255,255,0.07)", BgCardHover: "rgba(255,255,255,0.12)",
		TextPrimary: "#e8e8f0", TextSecondary: "#a3a9bc", TextMuted: "#7b82a0",
		BorderGlass: "rgba(255,255,255,0.12)", BorderGlassHover: "rgba(255,255,255,0.22)",
		TableStripe: "rgba(255,255,255,0.03)", InputBg: "rgba(255,255,255,0.05)",
		InputBgFocus: "rgba(255,255,255,0.08)", HoverBg: "rgba(255,255,255,0.04)",
		GlassWhite06: "rgba(255,255,255,0.07)", GlassWhite03: "rgba(255,255,255,0.04)",
		GlassWhite04: "rgba(255,255,255,0.05)", GlassWhite05: "rgba(255,255,255,0.06)",
		GlassWhite10: "rgba(255,255,255,0.12)", GlassWhite12: "rgba(255,255,255,0.14)",
		HeadingStart: "#ffffff", TextCode: "#d6e4f0",
	},
	"silver": {
		BgBody: "#e8eaf0", BgSidebar: "#dfe1e8", BgContent: "#f0f1f5",
		BgCode: "#e2e4ea", BgCard: "rgba(0,0,0,0.04)", BgCardHover: "rgba(0,0,0,0.07)",
		TextPrimary: "#1a1d2e", TextSecondary: "#4a5068", TextMuted: "#6b7394",
		BorderGlass: "rgba(0,0,0,0.10)", BorderGlassHover: "rgba(0,0,0,0.18)",
		TableStripe: "rgba(0,0,0,0.03)", InputBg: "rgba(0,0,0,0.04)",
		InputBgFocus: "rgba(0,0,0,0.06)", HoverBg: "rgba(0,0,0,0.03)",
		GlassWhite06: "rgba(0,0,0,0.04)", GlassWhite03: "rgba(0,0,0,0.02)",
		GlassWhite04: "rgba(0,0,0,0.03)", GlassWhite05: "rgba(0,0,0,0.04)",
		GlassWhite10: "rgba(0,0,0,0.08)", GlassWhite12: "rgba(0,0,0,0.10)",
		HeadingStart: "#1a1d2e", TextCode: "#374151",
	},
	"daylight": {
		BgBody: "#f8f9fc", BgSidebar: "#eef0f5", BgContent: "#ffffff",
		BgCode: "#f0f1f5", BgCard: "rgba(0,0,0,0.03)", BgCardHover: "rgba(0,0,0,0.06)",
		TextPrimary: "#111827", TextSecondary: "#4b5563", TextMuted: "#6b7280",
		BorderGlass: "rgba(0,0,0,0.08)", BorderGlassHover: "rgba(0,0,0,0.15)",
		TableStripe: "rgba(0,0,0,0.02)", InputBg: "rgba(0,0,0,0.03)",
		InputBgFocus: "rgba(0,0,0,0.05)", HoverBg: "rgba(0,0,0,0.02)",
		GlassWhite06: "rgba(0,0,0,0.03)", GlassWhite03: "rgba(0,0,0,0.02)",
		GlassWhite04: "rgba(0,0,0,0.03)", GlassWhite05: "rgba(0,0,0,0.03)",
		GlassWhite10: "rgba(0,0,0,0.06)", GlassWhite12: "rgba(0,0,0,0.08)",
		HeadingStart: "#111827", TextCode: "#1f2937",
	},
}

var accents = map[string]accentVars{
	"blue": {
		Accent1: "#2979ff", Accent2: "#00c6ff", Accent3: "#60a5fa", BtnGradEnd: "#5c9fff",
		AccentDim: "rgba(41,121,255,0.15)", GlowPurple: "rgba(41,121,255,0.20)",
		GlowBlue: "rgba(0,198,255,0.15)", HeadingTint: "#a8c8ff",
		BadgeBg: "rgba(41,121,255,0.12)", BadgeBorder: "rgba(41,121,255,0.25)",
		ActiveBg: "rgba(41,121,255,0.08)", HoverBg: "rgba(41,121,255,0.06)",
		FocusShadow: "rgba(41,121,255,0.15)",
		CardGrad1: "rgba(41,121,255,0.15)", CardGrad2: "rgba(0,198,255,0.1)",
		CardBorder: "rgba(41,121,255,0.2)",
		CardOverlay1: "rgba(41,121,255,0.08)", CardOverlay2: "rgba(0,198,255,0.04)",
		BtnShadow: "rgba(41,121,255,0.4)", AddCardHover: "rgba(41,121,255,0.06)",
		AddCardShadow: "rgba(41,121,255,0.1)",
		IconGrad2Bg: "rgba(0,198,255,0.15)", IconGrad2Bdr: "rgba(0,198,255,0.2)",
		IconGrad3Bg: "rgba(96,165,250,0.15)", IconGrad3Bdr: "rgba(96,165,250,0.2)",
		IconGradMix: "rgba(41,121,255,0.08)",
		EditHoverBg: "rgba(41,121,255,0.12)", VersionBdr: "rgba(41,121,255,0.2)",
		IconBoxShadow: "rgba(41,121,255,0.2)",
		TableHeadBg: "rgba(41,121,255,0.12)", TableHoverBg: "rgba(41,121,255,0.04)",
		BlockquoteBg: "rgba(41,121,255,0.06)", CopyHoverBdr: "rgba(41,121,255,0.3)",
	},
	"purple": {
		Accent1: "#7c3aed", Accent2: "#a78bfa", Accent3: "#c084fc", BtnGradEnd: "#9b6dff",
		AccentDim: "rgba(124,58,237,0.15)", GlowPurple: "rgba(124,58,237,0.20)",
		GlowBlue: "rgba(167,139,250,0.15)", HeadingTint: "#c4b5fd",
		BadgeBg: "rgba(124,58,237,0.12)", BadgeBorder: "rgba(124,58,237,0.25)",
		ActiveBg: "rgba(124,58,237,0.08)", HoverBg: "rgba(124,58,237,0.06)",
		FocusShadow: "rgba(124,58,237,0.15)",
		CardGrad1: "rgba(124,58,237,0.15)", CardGrad2: "rgba(167,139,250,0.1)",
		CardBorder: "rgba(124,58,237,0.2)",
		CardOverlay1: "rgba(124,58,237,0.08)", CardOverlay2: "rgba(167,139,250,0.04)",
		BtnShadow: "rgba(124,58,237,0.4)", AddCardHover: "rgba(124,58,237,0.06)",
		AddCardShadow: "rgba(124,58,237,0.1)",
		IconGrad2Bg: "rgba(167,139,250,0.15)", IconGrad2Bdr: "rgba(167,139,250,0.2)",
		IconGrad3Bg: "rgba(192,132,252,0.15)", IconGrad3Bdr: "rgba(192,132,252,0.2)",
		IconGradMix: "rgba(124,58,237,0.08)",
		EditHoverBg: "rgba(124,58,237,0.12)", VersionBdr: "rgba(124,58,237,0.2)",
		IconBoxShadow: "rgba(124,58,237,0.2)",
		TableHeadBg: "rgba(124,58,237,0.12)", TableHoverBg: "rgba(124,58,237,0.04)",
		BlockquoteBg: "rgba(124,58,237,0.06)", CopyHoverBdr: "rgba(124,58,237,0.3)",
	},
	"green": {
		Accent1: "#10b981", Accent2: "#34d399", Accent3: "#6ee7b7", BtnGradEnd: "#3dd68c",
		AccentDim: "rgba(16,185,129,0.15)", GlowPurple: "rgba(16,185,129,0.20)",
		GlowBlue: "rgba(52,211,153,0.15)", HeadingTint: "#a7f3d0",
		BadgeBg: "rgba(16,185,129,0.12)", BadgeBorder: "rgba(16,185,129,0.25)",
		ActiveBg: "rgba(16,185,129,0.08)", HoverBg: "rgba(16,185,129,0.06)",
		FocusShadow: "rgba(16,185,129,0.15)",
		CardGrad1: "rgba(16,185,129,0.15)", CardGrad2: "rgba(52,211,153,0.1)",
		CardBorder: "rgba(16,185,129,0.2)",
		CardOverlay1: "rgba(16,185,129,0.08)", CardOverlay2: "rgba(52,211,153,0.04)",
		BtnShadow: "rgba(16,185,129,0.4)", AddCardHover: "rgba(16,185,129,0.06)",
		AddCardShadow: "rgba(16,185,129,0.1)",
		IconGrad2Bg: "rgba(52,211,153,0.15)", IconGrad2Bdr: "rgba(52,211,153,0.2)",
		IconGrad3Bg: "rgba(110,231,183,0.15)", IconGrad3Bdr: "rgba(110,231,183,0.2)",
		IconGradMix: "rgba(16,185,129,0.08)",
		EditHoverBg: "rgba(16,185,129,0.12)", VersionBdr: "rgba(16,185,129,0.2)",
		IconBoxShadow: "rgba(16,185,129,0.2)",
		TableHeadBg: "rgba(16,185,129,0.12)", TableHoverBg: "rgba(16,185,129,0.04)",
		BlockquoteBg: "rgba(16,185,129,0.06)", CopyHoverBdr: "rgba(16,185,129,0.3)",
	},
	"orange": {
		Accent1: "#f59e0b", Accent2: "#fbbf24", Accent3: "#fcd34d", BtnGradEnd: "#f7b731",
		AccentDim: "rgba(245,158,11,0.15)", GlowPurple: "rgba(245,158,11,0.20)",
		GlowBlue: "rgba(251,191,36,0.15)", HeadingTint: "#fde68a",
		BadgeBg: "rgba(245,158,11,0.12)", BadgeBorder: "rgba(245,158,11,0.25)",
		ActiveBg: "rgba(245,158,11,0.08)", HoverBg: "rgba(245,158,11,0.06)",
		FocusShadow: "rgba(245,158,11,0.15)",
		CardGrad1: "rgba(245,158,11,0.15)", CardGrad2: "rgba(251,191,36,0.1)",
		CardBorder: "rgba(245,158,11,0.2)",
		CardOverlay1: "rgba(245,158,11,0.08)", CardOverlay2: "rgba(251,191,36,0.04)",
		BtnShadow: "rgba(245,158,11,0.4)", AddCardHover: "rgba(245,158,11,0.06)",
		AddCardShadow: "rgba(245,158,11,0.1)",
		IconGrad2Bg: "rgba(251,191,36,0.15)", IconGrad2Bdr: "rgba(251,191,36,0.2)",
		IconGrad3Bg: "rgba(252,211,77,0.15)", IconGrad3Bdr: "rgba(252,211,77,0.2)",
		IconGradMix: "rgba(245,158,11,0.08)",
		EditHoverBg: "rgba(245,158,11,0.12)", VersionBdr: "rgba(245,158,11,0.2)",
		IconBoxShadow: "rgba(245,158,11,0.2)",
		TableHeadBg: "rgba(245,158,11,0.12)", TableHoverBg: "rgba(245,158,11,0.04)",
		BlockquoteBg: "rgba(245,158,11,0.06)", CopyHoverBdr: "rgba(245,158,11,0.3)",
	},
	"red": {
		Accent1: "#ef4444", Accent2: "#f87171", Accent3: "#fca5a5", BtnGradEnd: "#f76c6c",
		AccentDim: "rgba(239,68,68,0.15)", GlowPurple: "rgba(239,68,68,0.20)",
		GlowBlue: "rgba(248,113,113,0.15)", HeadingTint: "#fecaca",
		BadgeBg: "rgba(239,68,68,0.12)", BadgeBorder: "rgba(239,68,68,0.25)",
		ActiveBg: "rgba(239,68,68,0.08)", HoverBg: "rgba(239,68,68,0.06)",
		FocusShadow: "rgba(239,68,68,0.15)",
		CardGrad1: "rgba(239,68,68,0.15)", CardGrad2: "rgba(248,113,113,0.1)",
		CardBorder: "rgba(239,68,68,0.2)",
		CardOverlay1: "rgba(239,68,68,0.08)", CardOverlay2: "rgba(248,113,113,0.04)",
		BtnShadow: "rgba(239,68,68,0.4)", AddCardHover: "rgba(239,68,68,0.06)",
		AddCardShadow: "rgba(239,68,68,0.1)",
		IconGrad2Bg: "rgba(248,113,113,0.15)", IconGrad2Bdr: "rgba(248,113,113,0.2)",
		IconGrad3Bg: "rgba(252,165,165,0.15)", IconGrad3Bdr: "rgba(252,165,165,0.2)",
		IconGradMix: "rgba(239,68,68,0.08)",
		EditHoverBg: "rgba(239,68,68,0.12)", VersionBdr: "rgba(239,68,68,0.2)",
		IconBoxShadow: "rgba(239,68,68,0.2)",
		TableHeadBg: "rgba(239,68,68,0.12)", TableHoverBg: "rgba(239,68,68,0.04)",
		BlockquoteBg: "rgba(239,68,68,0.06)", CopyHoverBdr: "rgba(239,68,68,0.3)",
	},
	"teal": {
		Accent1: "#14b8a6", Accent2: "#2dd4bf", Accent3: "#5eead4", BtnGradEnd: "#36cfc0",
		AccentDim: "rgba(20,184,166,0.15)", GlowPurple: "rgba(20,184,166,0.20)",
		GlowBlue: "rgba(45,212,191,0.15)", HeadingTint: "#99f6e4",
		BadgeBg: "rgba(20,184,166,0.12)", BadgeBorder: "rgba(20,184,166,0.25)",
		ActiveBg: "rgba(20,184,166,0.08)", HoverBg: "rgba(20,184,166,0.06)",
		FocusShadow: "rgba(20,184,166,0.15)",
		CardGrad1: "rgba(20,184,166,0.15)", CardGrad2: "rgba(45,212,191,0.1)",
		CardBorder: "rgba(20,184,166,0.2)",
		CardOverlay1: "rgba(20,184,166,0.08)", CardOverlay2: "rgba(45,212,191,0.04)",
		BtnShadow: "rgba(20,184,166,0.4)", AddCardHover: "rgba(20,184,166,0.06)",
		AddCardShadow: "rgba(20,184,166,0.1)",
		IconGrad2Bg: "rgba(45,212,191,0.15)", IconGrad2Bdr: "rgba(45,212,191,0.2)",
		IconGrad3Bg: "rgba(94,234,212,0.15)", IconGrad3Bdr: "rgba(94,234,212,0.2)",
		IconGradMix: "rgba(20,184,166,0.08)",
		EditHoverBg: "rgba(20,184,166,0.12)", VersionBdr: "rgba(20,184,166,0.2)",
		IconBoxShadow: "rgba(20,184,166,0.2)",
		TableHeadBg: "rgba(20,184,166,0.12)", TableHoverBg: "rgba(20,184,166,0.04)",
		BlockquoteBg: "rgba(20,184,166,0.06)", CopyHoverBdr: "rgba(20,184,166,0.3)",
	},
	"pink": {
		Accent1: "#ec4899", Accent2: "#f472b6", Accent3: "#f9a8d4", BtnGradEnd: "#f472b6",
		AccentDim: "rgba(236,72,153,0.15)", GlowPurple: "rgba(236,72,153,0.20)",
		GlowBlue: "rgba(244,114,182,0.15)", HeadingTint: "#fbcfe8",
		BadgeBg: "rgba(236,72,153,0.12)", BadgeBorder: "rgba(236,72,153,0.25)",
		ActiveBg: "rgba(236,72,153,0.08)", HoverBg: "rgba(236,72,153,0.06)",
		FocusShadow: "rgba(236,72,153,0.15)",
		CardGrad1: "rgba(236,72,153,0.15)", CardGrad2: "rgba(244,114,182,0.1)",
		CardBorder: "rgba(236,72,153,0.2)",
		CardOverlay1: "rgba(236,72,153,0.08)", CardOverlay2: "rgba(244,114,182,0.04)",
		BtnShadow: "rgba(236,72,153,0.4)", AddCardHover: "rgba(236,72,153,0.06)",
		AddCardShadow: "rgba(236,72,153,0.1)",
		IconGrad2Bg: "rgba(244,114,182,0.15)", IconGrad2Bdr: "rgba(244,114,182,0.2)",
		IconGrad3Bg: "rgba(249,168,212,0.15)", IconGrad3Bdr: "rgba(249,168,212,0.2)",
		IconGradMix: "rgba(236,72,153,0.08)",
		EditHoverBg: "rgba(236,72,153,0.12)", VersionBdr: "rgba(236,72,153,0.2)",
		IconBoxShadow: "rgba(236,72,153,0.2)",
		TableHeadBg: "rgba(236,72,153,0.12)", TableHoverBg: "rgba(236,72,153,0.04)",
		BlockquoteBg: "rgba(236,72,153,0.06)", CopyHoverBdr: "rgba(236,72,153,0.3)",
	},
}

// ValidTheme checks if a theme name is valid.
func ValidTheme(t string) bool {
	_, ok := themes[t]
	return ok
}

// ValidAccent checks if an accent color name is valid.
func ValidAccent(a string) bool {
	_, ok := accents[a]
	return ok
}

// ThemeCSS returns a <style> block that overrides :root CSS variables for the
// selected theme and accent color. For the default (midnight + blue) it returns
// an empty string so there is zero visual regression.
func ThemeCSS(themeName, accentColor string) template.HTML {
	if themeName == "" {
		themeName = "midnight"
	}
	if accentColor == "" {
		accentColor = "blue"
	}

	// Default combo: no override needed
	if themeName == "midnight" && accentColor == "blue" {
		return ""
	}

	t, ok := themes[themeName]
	if !ok {
		t = themes["midnight"]
	}
	a, ok := accents[accentColor]
	if !ok {
		a = accents["blue"]
	}

	css := fmt.Sprintf(`<style>
  :root {
    --bg-body: %s;
    --bg-sidebar: %s;
    --bg-content: %s;
    --bg-code: %s;
    --bg-card: %s;
    --bg-card-hover: %s;
    --text-primary: %s;
    --text-secondary: %s;
    --text-muted: %s;
    --border-glass: %s;
    --border-glass-hover: %s;
    --table-stripe: %s;
    --accent-1: %s;
    --accent-2: %s;
    --accent-3: %s;
    --accent-dim: %s;
    --glow-purple: %s;
    --glow-blue: %s;
    --btn-gradient-end: %s;
    --accent-heading-tint: %s;
    --accent-badge-bg: %s;
    --accent-badge-border: %s;
    --accent-active-bg: %s;
    --accent-hover-bg: %s;
    --accent-focus-shadow: %s;
    --accent-card-grad1: %s;
    --accent-card-grad2: %s;
    --accent-card-border: %s;
    --accent-card-overlay1: %s;
    --accent-card-overlay2: %s;
    --accent-btn-shadow: %s;
    --accent-add-card-hover: %s;
    --accent-add-card-shadow: %s;
    --accent-icon-grad2-bg: %s;
    --accent-icon-grad2-border: %s;
    --accent-icon-grad3-bg: %s;
    --accent-icon-grad3-border: %s;
    --accent-icon-grad-mix: %s;
    --accent-edit-hover-bg: %s;
    --accent-version-border: %s;
    --accent-icon-box-shadow: %s;
    --accent-table-head-bg: %s;
    --accent-table-hover-bg: %s;
    --accent-blockquote-bg: %s;
    --accent-copy-hover-border: %s;
    --heading-gradient-start: %s;
    --text-code: %s;
    --glass-white-06: %s;
    --glass-white-03: %s;
    --glass-white-04: %s;
    --glass-white-05: %s;
    --glass-white-10: %s;
    --glass-white-12: %s;
    --input-bg: %s;
    --input-bg-focus: %s;
    --hover-bg: %s;
  }
</style>`,
		t.BgBody, t.BgSidebar, t.BgContent, t.BgCode, t.BgCard, t.BgCardHover,
		t.TextPrimary, t.TextSecondary, t.TextMuted,
		t.BorderGlass, t.BorderGlassHover, t.TableStripe,
		a.Accent1, a.Accent2, a.Accent3, a.AccentDim,
		a.GlowPurple, a.GlowBlue, a.BtnGradEnd, a.HeadingTint,
		a.BadgeBg, a.BadgeBorder, a.ActiveBg, a.HoverBg, a.FocusShadow,
		a.CardGrad1, a.CardGrad2, a.CardBorder,
		a.CardOverlay1, a.CardOverlay2,
		a.BtnShadow, a.AddCardHover, a.AddCardShadow,
		a.IconGrad2Bg, a.IconGrad2Bdr, a.IconGrad3Bg, a.IconGrad3Bdr,
		a.IconGradMix, a.EditHoverBg, a.VersionBdr, a.IconBoxShadow,
		a.TableHeadBg, a.TableHoverBg, a.BlockquoteBg, a.CopyHoverBdr,
		t.HeadingStart, t.TextCode,
		t.GlassWhite06, t.GlassWhite03, t.GlassWhite04, t.GlassWhite05,
		t.GlassWhite10, t.GlassWhite12,
		t.InputBg, t.InputBgFocus, t.HoverBg,
	)
	return template.HTML(css)
}
