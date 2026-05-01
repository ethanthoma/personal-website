package layouts

import "personal-website/services/webserver/static"

// FontFacesInnerCSS holds just the @font-face rules, with no surrounding
// <style> tags. Exported so middleware can compute a stable CSP sha256 over
// the exact bytes between <style> and </style>; fontFacesHTML wraps it for
// templ rendering.
var FontFacesInnerCSS = buildFontFacesInnerCSS()
var fontFacesHTML = "<style>" + FontFacesInnerCSS + "</style>"

// Two @font-face declarations per fallback family, each calibrated against a
// different real fallback. The browser picks the first whose src local()
// chain resolves on the user's platform — so Mac/Windows users get
// Arial-class metrics, Linux users get DejaVu metrics, both with overrides
// tuned for the *actual* fallback rendering. This makes the LCP element's
// bounding box identical pre- and post-swap regardless of platform → no LCP
// candidate change at font load. Override values were computed from the
// shipped woff2 + the actual fallback ttf metrics (avgW per upem).
const (
	// Mac, Windows. Liberation Sans is metric-compatible with Arial so it
	// shares this declaration on Linux systems that have it.
	sansArialSrc       = `local("Arial"),local("Helvetica Neue"),local("Helvetica"),local("Liberation Sans")`
	sansArialAdjust    = "97.17%"
	sansArialAscent    = "97.77%"
	sansArialDescent   = "23.16%"
	sansArialLineGap   = "0%"
	// Linux fallback when Arial-class isn't installed.
	sansDejavuSrc      = `local("DejaVu Sans")`
	sansDejavuAdjust   = "85.75%"
	sansDejavuAscent   = "110.79%"
	sansDejavuDescent  = "26.24%"
	sansDejavuLineGap  = "0%"
	// All common monospace fallbacks land within 0.3% of each other for
	// Monaspace's metrics, so one declaration covers them.
	monoFallbackSrc    = `local("Menlo"),local("Consolas"),local("DejaVu Sans Mono"),local("Liberation Mono"),local("Courier New")`
	monoSizeAdjust     = "103.32%"
	monoAscent         = "91.47%"
	monoDescent        = "19.36%"
	monoLineGap        = "9.68%"
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
		fallback("Public Sans Fallback", sansArialSrc, sansArialAdjust, sansArialAscent, sansArialDescent, sansArialLineGap) +
		fallback("Public Sans Fallback", sansDejavuSrc, sansDejavuAdjust, sansDejavuAscent, sansDejavuDescent, sansDejavuLineGap) +
		fallback("Monaspace Fallback", monoFallbackSrc, monoSizeAdjust, monoAscent, monoDescent, monoLineGap)
}
