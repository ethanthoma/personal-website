(() => {
    const nav = { shown: location.pathname, pending: null, replaying: false };

    window.goTo = (path) => {
        nav.pending = path;
        if (!nav.replaying) {
            history.pushState(null, "", path);
            scrollTo(0, 0);
        }
        nav.shown = path;
    };

    addEventListener("popstate", () => {
        if (location.pathname === nav.shown) return;
        const link = document.querySelector(`[href="${location.pathname}"]`);
        if (!link) {
            location.reload();
            return;
        }
        nav.replaying = true;
        link.click();
        nav.replaying = false;
    });

    // If the fragment fetch dies (5xx, network drop, retries exhausted), fall
    // back to a full-page navigation so the user lands on the right URL with
    // fresh content instead of stale DOM.
    document.addEventListener("datastar-fetch", (evt) => {
        const t = evt.detail?.type;
        if (t === "finished") nav.pending = null;
        else if (nav.pending && (t === "error" || t === "retries-failed")) {
            location.href = nav.pending;
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

    function setFitCookie(name, k) {
        // biome-ignore lint/suspicious/noDocumentCookie: Cookie Store API lacks Firefox/Safari support
        document.cookie = `${name}=${k}; Path=/; SameSite=Lax; Max-Age=31536000`;
    }
    function countVisible(listId) {
        return document.querySelectorAll(`${listId} > li:not([hidden])`).length;
    }
    const overflowing = () =>
        document.documentElement.scrollHeight >
        (window.visualViewport?.height ?? window.innerHeight);

    function alternatingTail(projects, posts) {
        const tailOf = (list) =>
            list ? Array.from(list.children).slice(1).reverse() : [];
        const p = tailOf(projects);
        const w = tailOf(posts);
        const seq = [];
        for (let i = 0; i < Math.max(p.length, w.length); i++) {
            if (i < w.length) seq.push(w[i]);
            if (i < p.length) seq.push(p[i]);
        }
        return seq;
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

    const expandedLists = new Set();

    function fitHomeListsToViewport() {
        const projects = document.querySelector("#projects-list");
        const posts = document.querySelector("#posts-list");
        if (!projects && !posts) return;

        for (const list of [projects, posts]) {
            if (!list) continue;
            for (const li of list.children) li.hidden = false;
        }

        const collapsible = (list) =>
            list && !expandedLists.has(list.id) ? list : null;
        for (const li of alternatingTail(
            collapsible(projects),
            collapsible(posts),
        )) {
            if (!overflowing()) break;
            li.hidden = true;
        }

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
        expandedLists.add(list.id);
        for (const li of list.children) li.hidden = false;
        setFitCookie(
            isPosts ? "home-fit-posts" : "home-fit-projects",
            list.children.length,
        );
        btn.hidden = true;
    });

    let fitTimer = null;
    const debouncedFit = () => {
        clearTimeout(fitTimer);
        fitTimer = setTimeout(fitHomeListsToViewport, 50);
    };
    if (window.ResizeObserver) {
        new ResizeObserver(debouncedFit).observe(document.documentElement);
    } else {
        addEventListener("resize", debouncedFit);
    }
    window.visualViewport?.addEventListener("resize", debouncedFit);
    addEventListener("datastar-patch-elements", debouncedFit);
    document.fonts?.ready?.then(fitHomeListsToViewport);

    const prefetchedPostURLs = new Set();
    function prefetchPostFragment(slug) {
        const url = `/fragment/post/${slug}?${new URLSearchParams({ datastar: "{}" })}`;
        if (prefetchedPostURLs.has(url)) return;
        prefetchedPostURLs.add(url);
        fetch(url, { headers: { Accept: "text/event-stream" } })
            .then((r) => r.text())
            .catch(() => {});
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
    const prefetchPostFromEvent = (e) => {
        const slug = postSlugFromTarget(e.target);
        if (slug) prefetchPostFragment(slug);
    };
    addEventListener("pointerdown", prefetchPostFromEvent);
    addEventListener("focusin", prefetchPostFromEvent);
})();
