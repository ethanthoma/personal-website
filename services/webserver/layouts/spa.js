(() => {
	let displayedPath = location.pathname;
	let pendingPath = null;
	let inPopstate = false;

	window.goTo = (path) => {
		pendingPath = path;
		if (!inPopstate) {
			history.pushState(null, "", path);
			scrollTo(0, 0);
		}
		displayedPath = path;
	};

	addEventListener("popstate", () => {
		if (location.pathname === displayedPath) return;
		const link = document.querySelector(`[href="${location.pathname}"]`);
		if (!link) {
			location.reload();
			return;
		}
		inPopstate = true;
		link.click();
		inPopstate = false;
	});

	// If the fragment fetch dies (5xx, network drop, retries exhausted), fall
	// back to a full-page navigation so the user lands on the right URL with
	// fresh content instead of stale DOM.
	document.addEventListener("datastar-fetch", (evt) => {
		if (!pendingPath) return;
		const t = evt.detail?.type;
		if (t === "error" || t === "retries-failed") {
			location.href = pendingPath;
		}
	});

	let syncingHash = false;
	function syncDetailsToHash() {
		syncingHash = true;
		const hash = location.hash.slice(1);
		document.querySelectorAll("details").forEach((d) => {
			const content = d.querySelector("[id]");
			const shouldOpen = !!hash && content && content.id === hash;
			if (d.open !== shouldOpen) d.open = shouldOpen;
		});
		queueMicrotask(() =>
			queueMicrotask(() => {
				syncingHash = false;
			}),
		);
	}
	document.addEventListener(
		"toggle",
		(ev) => {
			if (syncingHash) return;
			if (ev.target.tagName !== "DETAILS") return;
			const content = ev.target.querySelector("[id]");
			if (!content) return;
			if (ev.target.open) {
				if (location.hash !== `#${content.id}`) {
					history.pushState(null, "", `#${content.id}`);
				}
			} else if (location.hash === `#${content.id}`) {
				history.pushState(null, "", location.pathname + location.search);
			}
		},
		true,
	);
	addEventListener("hashchange", syncDetailsToHash);
	addEventListener("datastar-patch-elements", syncDetailsToHash);
	syncDetailsToHash();

	function fitHomeListsToViewport() {
		const lists = [
			document.querySelector("#projects-list"),
			document.querySelector("#posts-list"),
		].filter(Boolean);
		if (!lists.length) return;

		for (const list of lists) {
			for (const li of list.querySelectorAll("li[data-trimmed]")) {
				li.style.display = "";
				li.removeAttribute("data-trimmed");
			}
		}
		for (const btn of document.querySelectorAll(
			"#show-all-posts, #show-all-projects",
		)) {
			btn.hidden = true;
		}

		const tailItems = [];
		for (let i = lists.length - 1; i >= 0; i--) {
			tailItems.push(...Array.from(lists[i].children).reverse());
		}
		for (const li of tailItems) {
			if (document.documentElement.scrollHeight <= window.innerHeight) break;
			li.style.display = "none";
			li.setAttribute("data-trimmed", "1");
		}

		const trimmedIn = (listId) =>
			!!document.querySelector(`${listId} li[data-trimmed]`);
		const postsBtn = document.querySelector("#show-all-posts");
		const projectsBtn = document.querySelector("#show-all-projects");
		if (postsBtn && trimmedIn("#posts-list")) postsBtn.hidden = false;
		if (projectsBtn && trimmedIn("#projects-list")) projectsBtn.hidden = false;
	}
	window.fitHomeListsToViewport = fitHomeListsToViewport;

	document.addEventListener("click", (e) => {
		const btn = e.target?.closest?.("#show-all-posts, #show-all-projects");
		if (!btn) return;
		const listId =
			btn.id === "show-all-posts" ? "#posts-list" : "#projects-list";
		for (const li of document.querySelectorAll(`${listId} li[data-trimmed]`)) {
			li.style.display = "";
			li.removeAttribute("data-trimmed");
		}
		btn.hidden = true;
	});

	addEventListener("datastar-patch-elements", fitHomeListsToViewport);
	if (document.readyState === "loading") {
		addEventListener("DOMContentLoaded", fitHomeListsToViewport);
	} else {
		fitHomeListsToViewport();
	}

	const prefetchedPostURLs = new Set();
	function prefetchPostFragment(slug) {
		const href = `/fragment/post/${slug}`;
		if (prefetchedPostURLs.has(href)) return;
		prefetchedPostURLs.add(href);
		const link = document.createElement("link");
		link.rel = "prefetch";
		link.as = "fetch";
		link.href = href;
		document.head.appendChild(link);
	}
	function postSlugFromTarget(t) {
		const a = t?.closest?.('a[href^="/post/"]');
		return a ? a.getAttribute("href").slice("/post/".length) : null;
	}
	let postHoverTimer = null;
	addEventListener("pointerover", (e) => {
		const slug = postSlugFromTarget(e.target);
		if (!slug) return;
		clearTimeout(postHoverTimer);
		postHoverTimer = setTimeout(() => prefetchPostFragment(slug), 80);
	});
	addEventListener("pointerout", () => clearTimeout(postHoverTimer));
	addEventListener("focusin", (e) => {
		const slug = postSlugFromTarget(e.target);
		if (slug) prefetchPostFragment(slug);
	});
})();
