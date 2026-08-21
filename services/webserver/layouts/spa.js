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

    // On fragment-fetch failure, full-page nav to the URL so the user gets fresh content, not stale DOM.
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

    // Email address is kept out of the served HTML to foil static scrapers;
    // assemble the mailto here, re-running after each fragment swap.
    function hydrateEmails() {
        document.querySelectorAll("a[data-email-user]").forEach((a) => {
            const { emailUser, emailDomain } = a.dataset;
            if (emailUser && emailDomain) {
                a.href = `mailto:${emailUser}@${emailDomain}`;
            }
        });
    }
    addEventListener("datastar-patch-elements", hydrateEmails);
    if (document.readyState === "loading") {
        addEventListener("DOMContentLoaded", hydrateEmails);
    } else {
        hydrateEmails();
    }

    document.addEventListener("click", (e) => {
        const btn = e.target?.closest?.("#show-all-posts, #show-all-projects");
        if (!btn) return;
        const list = document.querySelector(
            btn.id === "show-all-posts" ? "#posts-list" : "#projects-list",
        );
        if (list) for (const li of list.children) li.hidden = false;
        btn.hidden = true;
    });

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
