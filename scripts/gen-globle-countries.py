#! /usr/bin/env nix-shell
#! nix-shell -i python3 -p python3

"""Build the Globle country dataset from Natural Earth 50m borders.

Decodes the world-atlas TopoJSON into a compact per-country payload (name,
centroid, border rings). Simplification and rounding happen per shared ARC, not
per country ring, so neighbours keep byte-identical borders with no gap between.

Output: services/webserver/public/games/countries.json
"""

import json
import os
import urllib.request

SOURCE_URL = "https://cdn.jsdelivr.net/npm/world-atlas@2/countries-50m.json"
OUTPUT_PATH = os.path.join(
    os.path.dirname(__file__),
    "..",
    "services",
    "webserver",
    "public",
    "games",
    "countries.json",
)
COORD_DECIMALS = 2
SIMPLIFY_FACTOR = 0.05
SIMPLIFY_MAX = 0.15
SIMPLIFY_MIN = 0.01

DISPLAY_NAMES = {
    "United States of America": "United States",
    "Dem. Rep. Congo": "Democratic Republic of the Congo",
    "Congo": "Republic of the Congo",
    "Central African Rep.": "Central African Republic",
    "S. Sudan": "South Sudan",
    "Dominican Rep.": "Dominican Republic",
    "Bosnia and Herz.": "Bosnia and Herzegovina",
    "Eq. Guinea": "Equatorial Guinea",
    "Solomon Is.": "Solomon Islands",
    "Czechia": "Czech Republic",
    "Falkland Is.": "Falkland Islands",
    "Fr. S. Antarctic Lands": "French Southern and Antarctic Lands",
    "N. Cyprus": "Northern Cyprus",
    "W. Sahara": "Western Sahara",
    "eSwatini": "Eswatini",
    "Marshall Is.": "Marshall Islands",
    "Antigua and Barb.": "Antigua and Barbuda",
    "St. Kitts and Nevis": "Saint Kitts and Nevis",
    "St. Vin. and Gren.": "Saint Vincent and the Grenadines",
    "Macedonia": "North Macedonia",
}

EXCLUDED = {"Antarctica"}

SOURCE_10M_URL = "https://raw.githubusercontent.com/nvkelso/natural-earth-vector/master/geojson/ne_10m_admin_0_countries.geojson"

# Countries absent from the 50m base, pulled from Natural Earth's public-domain
# 10m set. Island nations only, so there is no shared-border gap with neighbours.
SUPPLEMENT_10M = {"Tuvalu"}


def ring_extent(points):
    xs = [p[0] for p in points]
    ys = [p[1] for p in points]
    return max(max(xs) - min(xs), max(ys) - min(ys))


def ring_area_centroid(ring):
    area = cx = cy = 0.0
    for i in range(len(ring)):
        x0, y0 = ring[i]
        x1, y1 = ring[(i + 1) % len(ring)]
        cross = x0 * y1 - x1 * y0
        area += cross
        cx += (x0 + x1) * cross
        cy += (y0 + y1) * cross
    if area == 0:
        return 0.0, ring[0]
    return abs(area / 2), (cx / (3 * area), cy / (3 * area))


def simplify(points, tolerance):
    n = len(points)
    if n < 3:
        return points
    keep = [False] * n
    keep[0] = keep[n - 1] = True
    tol2 = tolerance * tolerance
    stack = [(0, n - 1)]
    while stack:
        first, last = stack.pop()
        ax, ay = points[first]
        bx, by = points[last]
        dx, dy = bx - ax, by - ay
        seg2 = dx * dx + dy * dy
        max_d2, idx = 0.0, -1
        for i in range(first + 1, last):
            px, py = points[i]
            if seg2 == 0:
                d2 = (px - ax) ** 2 + (py - ay) ** 2
            else:
                t = max(0.0, min(1.0, ((px - ax) * dx + (py - ay) * dy) / seg2))
                d2 = (px - ax - t * dx) ** 2 + (py - ay - t * dy) ** 2
            if d2 > max_d2:
                max_d2, idx = d2, i
        if max_d2 > tol2 and idx != -1:
            keep[idx] = True
            stack.append((first, idx))
            stack.append((idx, last))
    return [points[i] for i in range(n) if keep[i]]


def build_arc_cache(topo):
    scale = topo["transform"]["scale"]
    translate = topo["transform"]["translate"]
    cache = []
    for arc in topo["arcs"]:
        x = y = 0
        points = []
        for dx, dy in arc:
            x += dx
            y += dy
            points.append((x * scale[0] + translate[0], y * scale[1] + translate[1]))
        tolerance = min(SIMPLIFY_MAX, max(SIMPLIFY_MIN, ring_extent(points) * SIMPLIFY_FACTOR))
        points = simplify(points, tolerance)
        cache.append(
            [[round(px, COORD_DECIMALS), round(py, COORD_DECIMALS)] for px, py in points]
        )
    return cache


def stitch_ring(arc_cache, ring_arc_indices):
    ring = []
    for index in ring_arc_indices:
        arc = arc_cache[~index][::-1] if index < 0 else arc_cache[index]
        ring.extend(arc[1:] if ring else arc)
    return ring


def clean_ring(ring):
    out = []
    for point in ring:
        if not out or out[-1] != point:
            out.append(point)
    if len(out) > 1 and out[0] == out[-1]:
        out.pop()
    return out


def polygons_of(geometry):
    if geometry["type"] == "Polygon":
        return [geometry["arcs"]]
    return geometry["arcs"]


def build_country(geometry, arc_cache):
    name = geometry.get("properties", {}).get("name", "")
    name = DISPLAY_NAMES.get(name, name)
    if name in EXCLUDED or not name:
        return None

    rings = []
    for polygon in polygons_of(geometry):
        for ring_arc_indices in polygon:
            ring = clean_ring(stitch_ring(arc_cache, ring_arc_indices))
            if len(ring) >= 2:
                rings.append(ring)
    if not rings:
        return None

    best_area, centroid = -1.0, None
    for ring in rings:
        if len(ring) >= 3:
            area, ring_centroid = ring_area_centroid(ring)
            if area > best_area:
                best_area, centroid = area, ring_centroid
    if centroid is None:
        biggest = max(rings, key=len)
        centroid = (
            sum(p[0] for p in biggest) / len(biggest),
            sum(p[1] for p in biggest) / len(biggest),
        )

    return {
        "n": name,
        "c": [round(centroid[0], 2), round(centroid[1], 2)],
        "p": [r for r in rings if len(r) >= 3],
    }


def outer_rings(geometry):
    if geometry["type"] == "Polygon":
        return [geometry["coordinates"][0]]
    if geometry["type"] == "MultiPolygon":
        return [polygon[0] for polygon in geometry["coordinates"]]
    return []


def build_supplement(name, geometry):
    rings = []
    for raw_ring in outer_rings(geometry):
        points = [(point[0], point[1]) for point in raw_ring]
        tolerance = min(SIMPLIFY_MAX, max(SIMPLIFY_MIN, ring_extent(points) * SIMPLIFY_FACTOR))
        points = simplify(points, tolerance)
        ring = clean_ring(
            [[round(px, COORD_DECIMALS), round(py, COORD_DECIMALS)] for px, py in points]
        )
        if len(ring) >= 3:
            rings.append(ring)
    if not rings:
        return None

    best_area, centroid = -1.0, None
    for ring in rings:
        area, ring_centroid = ring_area_centroid(ring)
        if area > best_area:
            best_area, centroid = area, ring_centroid
    if centroid is None:
        biggest = max(rings, key=len)
        centroid = (
            sum(p[0] for p in biggest) / len(biggest),
            sum(p[1] for p in biggest) / len(biggest),
        )

    return {
        "n": name,
        "c": [round(centroid[0], 2), round(centroid[1], 2)],
        "p": rings,
    }


def fetch_supplements():
    if not SUPPLEMENT_10M:
        return []
    with urllib.request.urlopen(SOURCE_10M_URL, timeout=60) as response:
        geojson = json.load(response)

    wanted = set(SUPPLEMENT_10M)
    supplements = []
    for feature in geojson["features"]:
        name = feature.get("properties", {}).get("NAME", "")
        name = DISPLAY_NAMES.get(name, name)
        if name in wanted:
            country = build_supplement(name, feature["geometry"])
            if country:
                supplements.append(country)
                wanted.discard(name)
    if wanted:
        raise SystemExit(f"10m supplement not found: {sorted(wanted)}")
    return supplements


def main():
    with urllib.request.urlopen(SOURCE_URL, timeout=30) as response:
        topo = json.load(response)

    arc_cache = build_arc_cache(topo)

    countries = []
    for geometry in topo["objects"]["countries"]["geometries"]:
        country = build_country(geometry, arc_cache)
        if country:
            countries.append(country)

    countries.extend(fetch_supplements())
    countries.sort(key=lambda c: c["n"])

    os.makedirs(os.path.dirname(OUTPUT_PATH), exist_ok=True)
    with open(OUTPUT_PATH, "w") as out:
        json.dump({"countries": countries}, out, separators=(",", ":"))

    size_kb = os.path.getsize(OUTPUT_PATH) / 1024
    invisible = sum(1 for c in countries if not c["p"])
    print(
        f"wrote {len(countries)} countries ({invisible} render-less specks), "
        f"{size_kb:.0f} KB -> {OUTPUT_PATH}"
    )


if __name__ == "__main__":
    main()
