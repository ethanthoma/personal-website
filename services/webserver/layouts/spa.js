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
                history.pushState(
                    null,
                    "",
                    location.pathname + location.search,
                );
            }
        },
        true,
    );
    addEventListener("hashchange", syncDetailsToHash);
    addEventListener("datastar-patch-elements", syncDetailsToHash);
    syncDetailsToHash();

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
