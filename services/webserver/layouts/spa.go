package layouts

// SpaRuntimeJS is the IIFE injected by spaRuntime(). Held as a Go constant so
// (a) middleware can compute a stable CSP sha256 over the exact bytes the
// browser sees, and (b) the inline <script> can drop 'unsafe-inline' from CSP
// without per-request nonces.
const SpaRuntimeJS = `(() => {
	let displayedPath = location.pathname;
	let pendingPath = null;
	let inPopstate = false;
	window.goTo = (path) => {
		pendingPath = path;
		if (!inPopstate) {
			history.pushState(null, '', path);
			scrollTo(0, 0);
		}
		displayedPath = path;
	};
	addEventListener("popstate", () => {
		if (location.pathname === displayedPath) return;
		const link = document.querySelector(` + "`" + `[href="${location.pathname}"]` + "`" + `);
		if (!link) { location.reload(); return; }
		inPopstate = true;
		link.click();
		inPopstate = false;
	});
	// If the fragment fetch dies (5xx, network drop, retries exhausted),
	// fall back to a full-page navigation so the user at least lands on
	// the right URL with fresh content instead of stale DOM.
	document.addEventListener("datastar-fetch", (evt) => {
		if (!pendingPath) return;
		const t = evt.detail && evt.detail.type;
		if (t === "error" || t === "retries-failed") {
			location.href = pendingPath;
		}
	});

	let syncingHash = false;
	function syncHash() {
		syncingHash = true;
		const hash = location.hash.slice(1);
		document.querySelectorAll("details").forEach(d => {
			const content = d.querySelector("[id]");
			const shouldOpen = !!hash && content && content.id === hash;
			if (d.open !== shouldOpen) d.open = shouldOpen;
		});
		queueMicrotask(() => queueMicrotask(() => { syncingHash = false; }));
	}
	document.addEventListener("toggle", (ev) => {
		if (syncingHash) return;
		if (ev.target.tagName !== "DETAILS") return;
		const content = ev.target.querySelector("[id]");
		if (!content) return;
		if (ev.target.open) {
			if (location.hash !== "#" + content.id) {
				history.pushState(null, '', "#" + content.id);
			}
		} else if (location.hash === "#" + content.id) {
			history.pushState(null, '', location.pathname + location.search);
		}
	}, true);
	addEventListener("hashchange", syncHash);
	addEventListener("datastar-patch-elements", syncHash);
	syncHash();
})();`
