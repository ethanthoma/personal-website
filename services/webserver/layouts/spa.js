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

	function setFitCookie(name, k) {
		// biome-ignore lint/suspicious/noDocumentCookie: Cookie Store API lacks Firefox/Safari support
		document.cookie = `${name}=${k}; Path=/; SameSite=Lax; Max-Age=31536000`;
	}
	function countVisible(listId) {
		return document.querySelectorAll(`${listId} > li:not([hidden])`).length;
	}
	const overflowing = () =>
		document.documentElement.scrollHeight > window.innerHeight;

	function trimTailUntilFits(projects, posts) {
		for (const list of [posts, projects]) {
			if (!list) continue;
			const items = list.children;
			for (let i = items.length - 1; i >= 1; i--) {
				if (!overflowing()) return;
				if (!items[i].hidden) items[i].hidden = true;
			}
		}
	}

	function expandUntilOverflow(projects, posts) {
		for (const list of [projects, posts]) {
			if (!list) continue;
			for (const item of list.children) {
				if (!item.hidden) continue;
				item.hidden = false;
				if (overflowing()) {
					item.hidden = true;
					return;
				}
			}
		}
	}

	function updateFitCookiesAndButtons(projects, posts) {
		const update = (list, cookieName, btnId) => {
			if (!list) return;
			const visible = Math.max(1, countVisible(`#${list.id}`));
			setFitCookie(cookieName, visible);
			const btn = document.querySelector(btnId);
			if (btn) btn.hidden = visible >= list.children.length;
		};
		update(projects, "home-fit-projects", "#show-all-projects");
		update(posts, "home-fit-posts", "#show-all-posts");
	}

	function ensureFirstItemVisible(list) {
		if (list?.children[0]?.hidden) list.children[0].hidden = false;
	}

	function fitHomeListsToViewport() {
		const projects = document.querySelector("#projects-list");
		const posts = document.querySelector("#posts-list");
		if (!projects && !posts) return;
		ensureFirstItemVisible(projects);
		ensureFirstItemVisible(posts);
		trimTailUntilFits(projects, posts);
		expandUntilOverflow(projects, posts);
		updateFitCookiesAndButtons(projects, posts);
	}
	window.fitHomeListsToViewport = fitHomeListsToViewport;

	document.addEventListener("click", (e) => {
		const btn = e.target?.closest?.("#show-all-posts, #show-all-projects");
		if (!btn) return;
		const isPosts = btn.id === "show-all-posts";
		const listId = isPosts ? "#posts-list" : "#projects-list";
		const list = document.querySelector(listId);
		if (!list) return;
		for (const li of list.children) li.hidden = false;
		setFitCookie(
			isPosts ? "home-fit-posts" : "home-fit-projects",
			list.children.length,
		);
		btn.hidden = true;
	});

	let resizeTimer = null;
	addEventListener("resize", () => {
		clearTimeout(resizeTimer);
		resizeTimer = setTimeout(fitHomeListsToViewport, 150);
	});
	addEventListener("datastar-patch-elements", fitHomeListsToViewport);
	document.fonts?.ready?.then(fitHomeListsToViewport);
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
