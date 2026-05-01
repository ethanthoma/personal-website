package layouts

import "personal-website/services/webserver/static"

// FontFacesInnerCSS holds just the @font-face rules, with no surrounding
// <style> tags. Exported so middleware can compute a stable CSP sha256 over
// the exact bytes between <style> and </style>; fontFacesHTML wraps it for
// templ rendering.
var FontFacesInnerCSS = buildFontFacesInnerCSS()
var fontFacesHTML = "<style>" + FontFacesInnerCSS + "</style>"

// Override values are computed from the shipped woff2 metrics divided by a
// reference fallback (Arial for sans, Courier New for mono). Glyph widths,
// ascent, descent, and line-gap are all forced to the target font's box, so a
// font-display: swap from fallback to real font produces zero box change → no
// LCP candidate change after first paint.
const (
	sansFallbackSrc  = `local("Arial"),local("Helvetica Neue"),local("Helvetica")`
	monoFallbackSrc  = `local("Menlo"),local("Consolas"),local("Liberation Mono"),local("Courier New")`
	sansSizeAdjust   = "103.32%"
	sansAscent       = "91.95%"
	sansDescent      = "21.78%"
	sansLineGap      = "0%"
	monoSizeAdjust   = "103.32%"
	monoAscent       = "91.46%"
	monoDescent      = "19.36%"
	monoLineGap      = "9.68%"
)

func buildFontFacesInnerCSS() string {
	neon := static.Asset("fonts/Monaspace/MonaspaceNeonVarVF[wght,wdth,slnt].woff2")
	krypton := static.Asset("fonts/Monaspace/MonaspaceKryptonVarVF[wght,wdth,slnt].woff2")
	publicSans := static.Asset("fonts/PublicSans/PublicSans[wght].woff2")
	publicSansItalic := static.Asset("fonts/PublicSans/PublicSans-Italic[wght].woff2")

	monaspace := func(family, url string) string {
		return `@font-face{font-family:"` + family + `";` +
			`src:url("` + url + `") format("woff2 supports variations"),` +
			`url("` + url + `") format("woff2-variations");` +
			`font-weight:200 800;font-stretch:100 125;` +
			`font-style:oblique -11deg 0deg;font-display:swap}`
	}
	publicSansFace := func(url, style string) string {
		return `@font-face{font-family:"Public Sans";` +
			`src:url("` + url + `") format("woff2 supports variations"),` +
			`url("` + url + `") format("woff2-variations");` +
			`font-weight:100 900;font-style:` + style + `;font-display:swap}`
	}
	fallback := func(family, src, sa, asc, desc, gap string) string {
		return `@font-face{font-family:"` + family + `";src:` + src +
			`;size-adjust:` + sa +
			`;ascent-override:` + asc +
			`;descent-override:` + desc +
			`;line-gap-override:` + gap + `}`
	}

	return monaspace("Monaspace Neon", neon) +
		monaspace("Monaspace Krypton", krypton) +
		publicSansFace(publicSans, "normal") +
		publicSansFace(publicSansItalic, "italic") +
		fallback("Public Sans Fallback", sansFallbackSrc, sansSizeAdjust, sansAscent, sansDescent, sansLineGap) +
		fallback("Monaspace Fallback", monoFallbackSrc, monoSizeAdjust, monoAscent, monoDescent, monoLineGap)
}
