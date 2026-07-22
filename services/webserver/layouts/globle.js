(() => {
    const root = document.getElementById("globle-root");
    const canvas = document.getElementById("globe");
    if (!root || !canvas) return;
    const ctx = canvas.getContext("2d");
    if (!ctx) return;

    const DEG = Math.PI / 180;
    const EARTH_RADIUS_KM = 6371;
    const WARMTH_FALLOFF_KM = 4000;
    const EPOCH_UTC = Date.UTC(2024, 0, 1);

    const cssVar = (name, fallback) =>
        getComputedStyle(document.documentElement)
            .getPropertyValue(name)
            .trim() || fallback;
    const parseHex = (hex) => {
        const h = hex.replace("#", "");
        const n = parseInt(h.length === 3 ? h.replace(/./g, "$&$&") : h, 16);
        return [(n >> 16) & 255, (n >> 8) & 255, n & 255];
    };
    const mix = (over, base, a) => {
        const o = parseHex(over),
            b = parseHex(base);
        const c = (i) => Math.round(o[i] * a + b[i] * (1 - a));
        return `rgb(${c(0)}, ${c(1)}, ${c(2)})`;
    };
    const CONTENT = cssVar("--color-content", "#141413");
    const BASE = cssVar("--color-base", "#FBF8EF");
    const OCEAN = mix(CONTENT, BASE, 0.05);
    const LAND = mix(CONTENT, BASE, 0.16);
    const STROKE = mix(CONTENT, BASE, 0.28);
    const FOUND = "oklch(0.62 0.15 150)";
    const REVEAL = cssVar("--color-blue", "#0064E6");
    const BORDER = CONTENT;

    const ALIASES = {
        unitedstates: ["usa", "us", "america"],
        unitedkingdom: ["uk", "britain", "greatbritain", "england", "gb"],
        unitedarabemirates: ["uae"],
        centralafricanrepublic: ["car"],
        democraticrepublicofthecongo: ["drc", "drcongo", "congokinshasa"],
        republicofthecongo: ["roc", "congo", "congobrazzaville"],
        dominicanrepublic: ["dr"],
        papuanewguinea: ["png"],
        bosniaandherzegovina: ["bih", "bosnia"],
        cotedivoire: ["ivorycoast"],
        caboverde: ["capeverde"],
        myanmar: ["burma"],
        czechrepublic: ["czechia"],
        eswatini: ["swaziland"],
        northmacedonia: ["macedonia"],
        antiguaandbarbuda: ["antigua"],
        saintkittsandnevis: ["stkittsandnevis", "stkitts"],
        saintvincentandthegrenadines: [
            "stvincent",
            "stvincentandthegrenadines",
        ],
        russia: ["russianfederation"],
        southkorea: ["korea", "rok", "republicofkorea"],
        northkorea: ["dprk"],
        netherlands: ["holland"],
    };

    const normalize = (s) =>
        s
            .normalize("NFD")
            .replace(/[̀-ͯ]/g, "")
            .toLowerCase()
            .replace(/[^a-z]/g, "");

    // Damerau-Levenshtein (adjacent transpositions cost 1, not 2) since swapped
    // letters are among the most common typos: "germnay" is one edit from Germany.
    function editDistance(a, b) {
        if (Math.abs(a.length - b.length) > 3) return 99;
        const d = Array.from({ length: a.length + 1 }, () =>
            new Array(b.length + 1).fill(0),
        );
        for (let i = 0; i <= a.length; i++) d[i][0] = i;
        for (let j = 0; j <= b.length; j++) d[0][j] = j;
        for (let i = 1; i <= a.length; i++) {
            for (let j = 1; j <= b.length; j++) {
                const cost = a[i - 1] === b[j - 1] ? 0 : 1;
                d[i][j] = Math.min(
                    d[i - 1][j] + 1,
                    d[i][j - 1] + 1,
                    d[i - 1][j - 1] + cost,
                );
                if (
                    i > 1 &&
                    j > 1 &&
                    a[i - 1] === b[j - 2] &&
                    a[i - 2] === b[j - 1]
                ) {
                    d[i][j] = Math.min(d[i][j], d[i - 2][j - 2] + 1);
                }
            }
        }
        return d[a.length][b.length];
    }
    function bestMatch(key) {
        let best = null;
        let otherDist = Infinity;
        for (const [k, c] of byKey) {
            const d = editDistance(key, k);
            if (!best || d < best.dist) {
                if (best && best.country !== c)
                    otherDist = Math.min(otherDist, best.dist);
                best = { country: c, dist: d };
            } else if (c !== best.country && d < otherDist) {
                otherDist = d;
            }
        }
        return best && { country: best.country, dist: best.dist, otherDist };
    }
    const toXYZ = (lng, lat) => {
        const l = lng * DEG,
            p = lat * DEG,
            c = Math.cos(p);
        return [c * Math.cos(l), c * Math.sin(l), Math.sin(p)];
    };
    const dot = (a, b) => a[0] * b[0] + a[1] * b[1] + a[2] * b[2];
    function haversineKm(a, b) {
        const p1 = a[1] * DEG,
            p2 = b[1] * DEG;
        const dp = (b[1] - a[1]) * DEG,
            dl = (b[0] - a[0]) * DEG;
        const x =
            Math.sin(dp / 2) ** 2 +
            Math.cos(p1) * Math.cos(p2) * Math.sin(dl / 2) ** 2;
        return 2 * EARTH_RADIUS_KM * Math.asin(Math.min(1, Math.sqrt(x)));
    }

    const distanceCache = new Map();
    function borderDistanceKm(a, b) {
        let min = Infinity;
        for (const ra of a.rings)
            for (const pa of ra)
                for (const rb of b.rings)
                    for (const pb of rb) {
                        const d = haversineKm(pa, pb);
                        if (d < min) min = d;
                    }
        return min;
    }
    function distanceKm(c) {
        if (c === target) return 0;
        let d = distanceCache.get(c);
        if (d === undefined) {
            d =
                c.rings.length && target.rings.length
                    ? borderDistanceKm(c, target)
                    : haversineKm(c.centroid, target.centroid);
            distanceCache.set(c, d);
        }
        return d;
    }

    function utcMidnight() {
        const now = new Date();
        return Date.UTC(
            now.getUTCFullYear(),
            now.getUTCMonth(),
            now.getUTCDate(),
        );
    }
    const dayIndex = Math.floor((utcMidnight() - EPOCH_UTC) / 86400000);
    const storeKey = `globle-${dayIndex}`;
    function hashInt(x) {
        x = Math.imul(x ^ 0x9e3779b9, 0x85ebca6b);
        x ^= x >>> 13;
        x = Math.imul(x, 0xc2b2ae35);
        x ^= x >>> 16;
        return x >>> 0;
    }

    let countries = [];
    const byKey = new Map();
    let target = null;
    let guesses = [];
    let won = false;
    let gaveUp = false;
    const view = { lng: 10, lat: 20 };

    function basis() {
        const l = view.lng * DEG,
            p = view.lat * DEG;
        const cp = Math.cos(p),
            sp = Math.sin(p),
            cl = Math.cos(l),
            sl = Math.sin(l);
        return {
            e: [-sl, cl, 0],
            n: [-sp * cl, -sp * sl, cp],
            f: [cp * cl, cp * sl, sp],
        };
    }

    function clipFront(ring, f) {
        const out = [];
        for (let i = 0; i < ring.length; i++) {
            const cur = ring[i],
                nxt = ring[(i + 1) % ring.length];
            const dc = dot(cur, f),
                dn = dot(nxt, f);
            if (dc >= 0) out.push(cur);
            if (dc >= 0 !== dn >= 0) {
                const t = dc / (dc - dn);
                out.push([
                    cur[0] + (nxt[0] - cur[0]) * t,
                    cur[1] + (nxt[1] - cur[1]) * t,
                    cur[2] + (nxt[2] - cur[2]) * t,
                ]);
            }
        }
        return out;
    }

    let cx = 0,
        cy = 0,
        R = 0,
        baseR = 0,
        dpr = 1,
        zoom = 1;
    const ZOOM_MIN = 1,
        ZOOM_MAX = 5;
    function setZoom(z) {
        zoom = Math.max(ZOOM_MIN, Math.min(ZOOM_MAX, z));
        R = baseR * zoom;
    }
    function resize() {
        const rect = canvas.getBoundingClientRect();
        if (!rect.width) return;
        dpr = window.devicePixelRatio || 1;
        canvas.width = Math.round(rect.width * dpr);
        canvas.height = Math.round(rect.height * dpr);
        cx = canvas.width / 2;
        cy = canvas.height / 2;
        baseR = Math.min(cx, cy) - Math.round(2 * dpr);
        R = baseR * zoom;
        draw();
    }

    function tracePath(country, b) {
        let any = false;
        ctx.beginPath();
        for (const ring of country.xyz) {
            const clip = clipFront(ring, b.f);
            const n = clip.length;
            if (n < 3) continue;
            const pts = clip.map((p) => ({
                X: cx + R * dot(p, b.e),
                Y: cy - R * dot(p, b.n),
                limb: dot(p, b.f) < 1e-6,
            }));
            ctx.moveTo(pts[0].X, pts[0].Y);
            for (let i = 1; i <= n; i++) {
                const prev = pts[i - 1],
                    cur = pts[i % n];
                if (prev.limb && cur.limb) {
                    const a0 = Math.atan2(prev.Y - cy, prev.X - cx);
                    const a1 = Math.atan2(cur.Y - cy, cur.X - cx);
                    let diff = a1 - a0;
                    while (diff > Math.PI) diff -= 2 * Math.PI;
                    while (diff < -Math.PI) diff += 2 * Math.PI;
                    ctx.arc(cx, cy, R, a0, a1, diff < 0);
                } else {
                    ctx.lineTo(cur.X, cur.Y);
                }
            }
            ctx.closePath();
            any = true;
        }
        return any;
    }

    function paintCountry(country, fill, stroke, b) {
        if (!tracePath(country, b)) return;
        ctx.fillStyle = fill;
        ctx.fill();
        ctx.lineWidth = (stroke === LAND ? 1 : 0.8) * dpr;
        ctx.strokeStyle = stroke;
        ctx.stroke();
    }

    function colorForDistance(km) {
        const warmth = Math.exp(-km / WARMTH_FALLOFF_KM);
        const lightness = 1 - 0.45 * warmth;
        const chroma = 0.19 * warmth;
        return `oklch(${lightness.toFixed(3)} ${chroma.toFixed(3)} 25)`;
    }

    function draw() {
        ctx.clearRect(0, 0, canvas.width, canvas.height);
        ctx.beginPath();
        ctx.arc(cx, cy, baseR, 0, 2 * Math.PI);
        ctx.fillStyle = OCEAN;
        ctx.fill();

        ctx.save();
        ctx.beginPath();
        ctx.arc(cx, cy, baseR, 0, 2 * Math.PI);
        ctx.clip();
        const b = basis();
        const guessed = new Set(guesses);
        for (const c of countries) {
            if (guessed.has(c)) continue;
            paintCountry(c, LAND, LAND, b);
        }
        for (const c of guesses) {
            if (c === target) continue;
            paintCountry(c, colorForDistance(distanceKm(c)), BORDER, b);
        }
        if ((won || gaveUp) && target)
            paintCountry(target, won ? FOUND : REVEAL, BORDER, b);
        ctx.restore();

        ctx.beginPath();
        ctx.arc(cx, cy, baseR, 0, 2 * Math.PI);
        ctx.lineWidth = dpr;
        ctx.strokeStyle = STROKE;
        ctx.stroke();
    }

    let rafPending = false;
    function scheduleDraw() {
        if (rafPending) return;
        rafPending = true;
        requestAnimationFrame(() => {
            rafPending = false;
            draw();
        });
    }

    function animateTo(lng, lat, zoomTarget) {
        cancelInertia();
        const startLng = view.lng,
            startLat = view.lat;
        const startZoom = zoom;
        const endZoom =
            zoomTarget === undefined
                ? zoom
                : Math.max(ZOOM_MIN, Math.min(ZOOM_MAX, zoomTarget));
        const dLng = ((lng - startLng + 540) % 360) - 180;
        const dLat = lat - startLat;
        const t0 = performance.now();
        (function step(now) {
            const k = Math.min(1, (now - t0) / 500);
            const e = 1 - (1 - k) ** 3;
            view.lng = startLng + dLng * e;
            view.lat = startLat + dLat * e;
            if (zoomTarget !== undefined)
                setZoom(startZoom + (endZoom - startZoom) * e);
            draw();
            if (k < 1) requestAnimationFrame(step);
        })(t0);
    }

    const message = (text) => {
        document.getElementById("globle-message").textContent = text;
    };

    function renderGuesses() {
        const list = document.getElementById("globle-guesses");
        list.replaceChildren();
        const rows = guesses
            .map((c, i) => ({ c, i, km: distanceKm(c) }))
            .sort((a, b) => a.km - b.km || b.i - a.i);
        for (const { c, km } of rows) {
            const li = document.createElement("li");
            li.className =
                "flex items-center gap-xs cursor-pointer hover:text-blue";
            li.addEventListener("click", () =>
                animateTo(c.centroid[0], c.centroid[1], 2.5),
            );
            const swatch = document.createElement("span");
            swatch.className = "inline-block shrink-0";
            swatch.style.cssText = `width:0.9rem;height:0.9rem;border:1px solid ${BORDER};`;
            swatch.style.background =
                c === target ? FOUND : colorForDistance(km);
            const name = document.createElement("span");
            name.className = "flex-1 min-w-0 font-sans";
            name.textContent = c.name;
            const dist = document.createElement("span");
            dist.className = "tabular-nums text-content/60";
            dist.textContent =
                c === target
                    ? "found"
                    : `${Math.round(km).toLocaleString()} km`;
            li.append(swatch, name, dist);
            list.append(li);
        }
    }

    function squareFor(km) {
        if (km < 1500) return "🟥";
        if (km < 4000) return "🟧";
        if (km < 8000) return "🟨";
        if (km < 13000) return "🟦";
        return "⬜";
    }
    function shareText() {
        const squares = guesses
            .map((c) => (c === target ? "🟩" : squareFor(distanceKm(c))))
            .join("");
        const date = new Date(utcMidnight()).toISOString().slice(0, 10);
        return `Globle ${date}\n${squares}\nGuessed in ${guesses.length}\n${location.origin}/games/globle`;
    }

    function loadStats() {
        try {
            return JSON.parse(localStorage.getItem("globle-stats")) || {};
        } catch (e) {
            return {};
        }
    }
    function recordWin() {
        const s = loadStats();
        if (s.lastWinDay !== dayIndex) {
            s.streak = s.lastWinDay === dayIndex - 1 ? (s.streak || 0) + 1 : 1;
            s.best = Math.max(s.best || 0, s.streak);
            s.wins = (s.wins || 0) + 1;
            s.lastWinDay = dayIndex;
            try {
                localStorage.setItem("globle-stats", JSON.stringify(s));
            } catch (e) {}
        }
        return s;
    }
    function recordGiveUp() {
        const s = loadStats();
        if (s.lastGiveUpDay !== dayIndex) {
            s.streak = 0;
            s.lastGiveUpDay = dayIndex;
            try {
                localStorage.setItem("globle-stats", JSON.stringify(s));
            } catch (e) {}
        }
    }

    function nextPuzzleCountdown() {
        const now = new Date();
        const ms =
            Date.UTC(
                now.getUTCFullYear(),
                now.getUTCMonth(),
                now.getUTCDate() + 1,
            ) - now.getTime();
        const pad = (n) => String(n).padStart(2, "0");
        return `${Math.floor(ms / 3600000)}h ${pad(Math.floor(ms / 60000) % 60)}m ${pad(Math.floor(ms / 1000) % 60)}s`;
    }
    function startCountdown() {
        const el = document.getElementById("globle-countdown");
        const tick = () => {
            el.textContent = nextPuzzleCountdown();
        };
        tick();
        setInterval(tick, 1000);
    }

    function lockSolved() {
        const input = document.getElementById("guess-input");
        const button = document.querySelector("#guess-form button");
        const plural = guesses.length === 1 ? "guess" : "guesses";
        input.value = "";
        input.disabled = true;
        input.placeholder = gaveUp
            ? `Answer: ${target.name}`
            : `Solved in ${guesses.length} ${plural}`;
        input.classList.add("text-center", "cursor-default");
        input.classList.add(gaveUp ? "placeholder:text-blue" : "text-content");
        if (button) button.hidden = true;
        const giveUpLink = document.getElementById("give-up");
        if (giveUpLink) giveUpLink.classList.add("invisible");
    }

    function showResult() {
        const el = document.getElementById("globle-message");
        el.replaceChildren();
        const squares = document.createElement("span");
        squares.style.cssText =
            "display:inline-flex;vertical-align:middle;gap:2px;margin-right:0.35rem;";
        for (const c of guesses) {
            const sq = document.createElement("span");
            sq.style.cssText = `display:inline-block;width:0.85rem;height:0.85rem;border:1px solid ${BORDER};`;
            sq.style.background =
                c === target ? FOUND : colorForDistance(distanceKm(c));
            squares.append(sq);
        }
        const copy = document.createElement("button");
        copy.type = "button";
        copy.textContent = "copy";
        copy.className =
            "underline underline-offset-2 cursor-pointer hover:text-blue";
        copy.addEventListener("click", async () => {
            try {
                await navigator.clipboard.writeText(shareText());
                copy.textContent = "copied";
                setTimeout(() => {
                    copy.textContent = "copy";
                }, 1500);
            } catch (e) {}
        });
        el.append(squares, copy);
        const stats = loadStats();
        if (stats.streak > 1) {
            const streak = document.createElement("span");
            streak.className = "ml-xs text-content/60";
            streak.textContent = `${stats.streak} day streak 🔥`;
            el.append(streak);
        }
    }

    function finishWin() {
        won = true;
        recordWin();
        showResult();
        lockSolved();
    }

    function finishGiveUp() {
        gaveUp = true;
        recordGiveUp();
        save();
        draw();
        message(`The answer was ${target.name}.`);
        document.getElementById("globle-message").classList.add("invisible");
        lockSolved();
    }

    function submitGuess(raw) {
        if (won || gaveUp) return;
        const key = normalize(raw);
        let country = byKey.get(key);
        if (!country && key) {
            const m = bestMatch(key);
            const short = key.length <= 5;
            const forgives =
                m &&
                m.dist <= (short ? 1 : 2) &&
                m.dist < m.otherDist &&
                m.dist <= Math.floor(key.length / 3);
            if (forgives) {
                country = m.country;
            } else if (m && m.dist <= 3 && m.dist * 2 <= key.length) {
                message(`Unknown country. Did you mean ${m.country.name}?`);
                return;
            }
        }
        if (!country) {
            message(`Unknown country: "${raw.trim()}"`);
            return;
        }
        if (guesses.includes(country)) {
            message(`Already guessed ${country.name}.`);
            animateTo(country.centroid[0], country.centroid[1]);
            return;
        }
        guesses.push(country);
        save();
        renderGuesses();
        animateTo(country.centroid[0], country.centroid[1]);
        if (country === target) {
            finishWin();
        } else {
            const km = distanceKm(country);
            message(
                km < 50
                    ? `${country.name} is right next door!`
                    : `${country.name} is ${Math.round(km).toLocaleString()} km away.`,
            );
        }
    }

    function save() {
        try {
            localStorage.setItem(
                storeKey,
                JSON.stringify({
                    guesses: guesses.map((c) => c.name),
                    won,
                    gaveUp,
                }),
            );
        } catch (e) {}
    }
    function restore() {
        let state;
        try {
            state = JSON.parse(localStorage.getItem(storeKey) || "null");
        } catch (e) {
            return;
        }
        if (!state) return;
        for (const name of state.guesses || []) {
            const c = byKey.get(normalize(name));
            if (c && !guesses.includes(c)) guesses.push(c);
        }
        won = guesses.includes(target);
        gaveUp = !won && Boolean(state.gaveUp);
    }

    const clampLat = (lat) => Math.max(-90, Math.min(90, lat));

    let inertiaRAF = 0;
    let velLng = 0,
        velLat = 0;
    function cancelInertia() {
        velLng = velLat = 0;
        if (inertiaRAF) {
            cancelAnimationFrame(inertiaRAF);
            inertiaRAF = 0;
        }
    }
    function startInertia() {
        let last = performance.now();
        const tick = (now) => {
            // rAF's timestamp can predate the performance.now() captured in the
            // pointerup handler, making a raw (now - last) negative and killing the
            // glide on frame one. Clamp to a sane positive frame delta.
            const dt = Math.min(64, Math.max(1, now - last));
            last = now;
            view.lng += velLng * dt;
            view.lat = clampLat(view.lat + velLat * dt);
            const decay = 0.9 ** (dt / 16);
            velLng *= decay;
            velLat *= decay;
            draw();
            inertiaRAF =
                Math.hypot(velLng, velLat) > 0.001
                    ? requestAnimationFrame(tick)
                    : 0;
        };
        inertiaRAF = requestAnimationFrame(tick);
    }

    const pointers = new Map();
    let dragging = false;
    let lastX = 0,
        lastY = 0,
        lastMove = 0;
    let pinchDist = 0,
        pinchZoom = 1;

    canvas.addEventListener("pointerdown", (e) => {
        cancelInertia();
        try {
            canvas.setPointerCapture(e.pointerId);
        } catch (err) {}
        pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
        if (pointers.size === 1) {
            dragging = true;
            lastX = e.clientX;
            lastY = e.clientY;
            lastMove = performance.now();
            canvas.style.cursor = "grabbing";
        } else if (pointers.size === 2) {
            dragging = false;
            const [a, b] = [...pointers.values()];
            pinchDist = Math.hypot(a.x - b.x, a.y - b.y);
            pinchZoom = zoom;
        }
    });

    canvas.addEventListener("pointermove", (e) => {
        if (!pointers.has(e.pointerId)) return;
        pointers.set(e.pointerId, { x: e.clientX, y: e.clientY });
        if (pointers.size >= 2) {
            const [a, b] = [...pointers.values()];
            const d = Math.hypot(a.x - b.x, a.y - b.y);
            if (pinchDist > 0) setZoom(pinchZoom * (d / pinchDist));
            scheduleDraw();
            return;
        }
        if (!dragging) return;
        const perPx = ((180 / Math.PI) * dpr) / R;
        const cosLat = Math.max(Math.cos(view.lat * DEG), 0.2);
        const dLng = (-(e.clientX - lastX) * perPx) / cosLat;
        const dLat = (e.clientY - lastY) * perPx;
        view.lng += dLng;
        view.lat = clampLat(view.lat + dLat);
        const now = performance.now();
        const dt = Math.max(1, now - lastMove);
        const w = Math.exp(-dt / 1000);
        velLng = w * (dLng / dt) + (1 - w) * velLng;
        velLat = w * (dLat / dt) + (1 - w) * velLat;
        lastX = e.clientX;
        lastY = e.clientY;
        lastMove = now;
        scheduleDraw();
    });

    function releasePointer(e) {
        if (!pointers.has(e.pointerId)) return;
        pointers.delete(e.pointerId);
        if (pointers.size >= 2) return;
        if (pointers.size === 1) {
            const p = [...pointers.values()][0];
            dragging = true;
            lastX = p.x;
            lastY = p.y;
            lastMove = performance.now();
            velLng = velLat = 0;
            return;
        }
        if (!dragging) return;
        dragging = false;
        canvas.style.cursor = "grab";
        const idle = performance.now() - lastMove;
        const speed = Math.hypot(velLng, velLat);
        if (idle < 120 && speed > 0.002) {
            const VMAX = 0.3;
            if (speed > VMAX) {
                velLng *= VMAX / speed;
                velLat *= VMAX / speed;
            }
            startInertia();
        } else {
            velLng = velLat = 0;
        }
    }
    canvas.addEventListener("pointerup", releasePointer);
    canvas.addEventListener("pointercancel", releasePointer);

    canvas.addEventListener(
        "wheel",
        (e) => {
            e.preventDefault();
            cancelInertia();
            setZoom(zoom * Math.exp(-e.deltaY * 0.0015));
            scheduleDraw();
        },
        { passive: false },
    );

    const hoverLabel = document.createElement("div");
    hoverLabel.style.cssText = `position:fixed;pointer-events:none;z-index:50;display:none;padding:1px 6px;border:1px solid ${STROKE};background:${BASE};color:${CONTENT};font-size:0.72rem;white-space:nowrap;`;
    document.body.appendChild(hoverLabel);

    function pointerLngLat(e) {
        const rect = canvas.getBoundingClientRect();
        const x = ((e.clientX - rect.left) * dpr - cx) / R;
        const y = -(((e.clientY - rect.top) * dpr - cy) / R);
        const d2 = x * x + y * y;
        if (d2 > 1) return null;
        const z = Math.sqrt(1 - d2);
        const b = basis();
        const wx = x * b.e[0] + y * b.n[0] + z * b.f[0];
        const wy = x * b.e[1] + y * b.n[1] + z * b.f[1];
        const wz = x * b.e[2] + y * b.n[2] + z * b.f[2];
        return [
            Math.atan2(wy, wx) / DEG,
            Math.asin(Math.max(-1, Math.min(1, wz))) / DEG,
        ];
    }
    function pointInRing(lng, lat, ring) {
        let inside = false;
        for (let i = 0, j = ring.length - 1; i < ring.length; j = i++) {
            const xi = ring[i][0],
                yi = ring[i][1];
            const xj = ring[j][0],
                yj = ring[j][1];
            if (
                yi > lat !== yj > lat &&
                lng < ((xj - xi) * (lat - yi)) / (yj - yi) + xi
            )
                inside = !inside;
        }
        return inside;
    }
    function countryAt(lng, lat) {
        for (const c of guesses)
            for (const ring of c.rings)
                if (pointInRing(lng, lat, ring)) return c;
        if (gaveUp && target)
            for (const ring of target.rings)
                if (pointInRing(lng, lat, ring)) return target;
        return null;
    }
    canvas.addEventListener("pointermove", (e) => {
        if (dragging || pointers.size > 0) {
            hoverLabel.style.display = "none";
            return;
        }
        const ll = pointerLngLat(e);
        const c = ll && countryAt(ll[0], ll[1]);
        if (c) {
            hoverLabel.textContent = c.name;
            hoverLabel.style.left = `${e.clientX + 12}px`;
            hoverLabel.style.top = `${e.clientY + 12}px`;
            hoverLabel.style.display = "block";
            canvas.style.cursor = "pointer";
        } else {
            hoverLabel.style.display = "none";
            canvas.style.cursor = "grab";
        }
    });
    canvas.addEventListener("pointerleave", () => {
        hoverLabel.style.display = "none";
    });

    document.getElementById("guess-form").addEventListener("submit", (e) => {
        e.preventDefault();
        const input = document.getElementById("guess-input");
        if (input.value.trim()) submitGuess(input.value);
        input.value = "";
    });

    const giveUpDialog = document.getElementById("giveup-dialog");
    document.getElementById("give-up").addEventListener("click", () => {
        if (won || gaveUp) return;
        const streak = loadStats().streak || 0;
        document.getElementById("giveup-text").textContent =
            streak >= 1
                ? `Give up and reveal today's answer? This ends your ${streak} day streak 🔥.`
                : "Give up and reveal today's answer?";
        giveUpDialog.showModal();
    });
    document
        .getElementById("giveup-cancel")
        .addEventListener("click", () => giveUpDialog.close());
    document.getElementById("giveup-confirm").addEventListener("click", () => {
        giveUpDialog.close();
        finishGiveUp();
        animateTo(target.centroid[0], target.centroid[1]);
    });
    if (window.ResizeObserver)
        new ResizeObserver(() => resize()).observe(canvas);
    else window.addEventListener("resize", resize);

    async function init() {
        let data;
        try {
            const res = await fetch(root.dataset.countries, {
                headers: { Accept: "application/json" },
            });
            data = await res.json();
        } catch (e) {
            message("Couldn't load the map. Try refreshing.");
            return;
        }
        countries = data.countries.map((c) => ({
            name: c.n,
            centroid: c.c,
            rings: c.p,
            xyz: c.p.map((ring) => ring.map(([lng, lat]) => toXYZ(lng, lat))),
        }));
        for (const c of countries) {
            byKey.set(normalize(c.name), c);
            for (const alias of ALIASES[normalize(c.name)] || [])
                byKey.set(alias, c);
        }
        target = countries[hashInt(dayIndex) % countries.length];

        restore();
        renderGuesses();
        if (won) {
            finishWin();
            animateTo(target.centroid[0], target.centroid[1]);
        } else if (gaveUp) {
            finishGiveUp();
            animateTo(target.centroid[0], target.centroid[1]);
        } else if (guesses.length) {
            const nearest = guesses.reduce((a, c) =>
                distanceKm(c) < distanceKm(a) ? c : a,
            );
            animateTo(nearest.centroid[0], nearest.centroid[1]);
        }
        resize();
        startCountdown();
    }
    init();
})();
