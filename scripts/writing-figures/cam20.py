"""Figures for the cam20 pack (our own drawings and data)."""
from svgkit import ARROW_DEFS, INK, MUTED, arrow, badge, box, lines, pie, rect, svg, table, text, vbars, wrap

WATER = "#cfe0f0"
FOREST = "#cfe3c2"
GRASS = "#f1f6ea"


def population_tables(f):
    w = 720
    out = []
    y = 40
    for t in f["tables"]:
        cols = [180, 240] if len(t["columns"]) == 2 else [120, 200, 300]
        x0 = (w - sum(cols)) / 2
        frag, y = table(t["title"], t["columns"], t["rows"], x0, y, cols)
        out += frag
        y += 52
    return svg(w, y - 20, "\n".join(out))


def compass(x, y):
    return (f'<circle cx="{x}" cy="{y - 6}" r="36" fill="#ffffff" stroke="#d9dee3"/>'
            f'<path d="M{x},{y - 22} L{x},{y + 22} M{x - 22},{y} L{x + 22},{y}" stroke="{INK}" stroke-width="1.4"/>'
            f'<path d="M{x},{y - 30} l-5,10 h10 z" fill="{INK}"/>' + text(x, y - 34, "N", 13, "middle", "bold"))


def trees(x, y, cols, rows, gap=26):
    out = []
    for r in range(rows):
        for c in range(cols):
            tx, ty = x + c * gap + (r % 2) * gap / 2, y + r * gap
            out.append(f'<path d="M{tx},{ty - 10} l8,14 h-16 z" fill="#548235"/>')
    return out


def forest_map(ox, oy, label, today):
    W, H = 680, 340
    out = [text(ox, oy - 12, label, 17, "start", "bold"), rect(ox, oy, W, H, GRASS, MUTED, 1.4)]
    # main road (north)
    out.append(f'<rect x="{ox}" y="{oy}" width="{W}" height="26" fill="#cbd2d9"/>')
    out.append(text(ox + W / 2, oy + 18, "Main road", 12, "middle"))
    # forest blocks: east always; west only in 1990
    out.append(rect(ox + 450, oy + 40, 210, 280, FOREST, "#8fb37a", 1.2, 12))
    out += trees(ox + 476, oy + 70, 7, 9)
    if not today:
        out.append(rect(ox + 20, oy + 120, 210, 200, FOREST, "#8fb37a", 1.2, 12))
        out += trees(ox + 44, oy + 150, 7, 6)
    # lake + stream
    cx, cy = ox + 340, oy + 175
    out.append(f'<path d="M{cx},{cy + 80} C{cx + 10},{cy + 120} {cx - 20},{cy + 140} {cx},{oy + H}" fill="none" stroke="#7fa7c9" stroke-width="7"/>')
    out.append(text(cx + 14, oy + H - 12, "Stream", 12, "start", fill="#1f4e79"))
    if today:
        out.append(f'<ellipse cx="{cx}" cy="{cy}" rx="118" ry="96" fill="none" stroke="#c55a11" stroke-width="2.6" stroke-dasharray="10 6"/>')
        out.append(text(cx - 125, cy + 118, "Cycle trail", 12, "start", "bold", "#c55a11"))
    out.append(f'<ellipse cx="{cx}" cy="{cy}" rx="95" ry="76" fill="{WATER}" stroke="#7fa7c9" stroke-width="1.5"/>')
    out.append(text(cx, cy + 6, "Lake", 15, "middle", "bold", "#1f4e79"))
    # car park and the ranger's hut / toilets
    if today:
        out.append(box(ox + 20, oy + 38, 200, 56, ["Car park"]))
        out.append(box(ox + 232, oy + 38, 70, 40, ["Toilets"], size=12))
        out.append(box(ox + 30, oy + 120, 130, 56, ["Visitor centre", "& café"], "#fff7e6", 13))
        out.append(f'<rect x="{ox + 30}" y="{oy + 220}" width="160" height="100" rx="8" fill="#ffffff" stroke="{MUTED}" stroke-dasharray="5 4"/>')
        for i in range(3):
            tx = ox + 60 + i * 46
            out.append(f'<path d="M{tx},{oy + 290} l14,-24 l14,24 z" fill="#ffffff" stroke="{INK}" stroke-width="1.4"/>')
        out.append(text(ox + 110, oy + 312, "Campsite", 13, "middle"))
        out.append(rect(cx - 104, cy - 14, 22, 10, "#a0805a", INK, 1))
        out.append(text(cx - 168, cy + 14, "Boat hire", 12, "start"))
        out.append(rect(cx + 60, cy + 46, 30, 8, "#a0805a", INK, 1))
        out.append(text(cx + 30, cy + 74, "Fishing platform", 11, "start"))
    else:
        out.append(box(ox + 20, oy + 38, 70, 56, ["Car", "park"]))
        out.append(box(ox + 100, oy + 38, 92, 40, ["Ranger's hut"], size=12))
    # footpath (both maps)
    out.append(f'<path d="M{ox + 90},{oy + 96} C{ox + 160},{oy + 110} {ox + 200},{oy + 150} {cx - 92},{cy - 8}" fill="none" stroke="{MUTED}" stroke-width="2" stroke-dasharray="3 5"/>')
    out.append(text(ox + 236, oy + 128, "Footpath", 12, "start", fill=MUTED))
    out.append(compass(ox + W - 40, oy + 70))
    return out


def museum(f):
    w, h = 820, 780
    out = pie(f["pie"], 190, 200, 120, 40, 360, title_y=46, title_x=200)
    frag, _ = table(f["table"]["title"], [""] + f["table"]["columns"],
                    [[r["label"]] + r["values"] for r in f["table"]["rows"]], 430, 46, [220, 140])
    out += frag
    out += vbars(f["bars"], 40, 500, 740, 270)
    return svg(w, h, "\n".join(out))


def recycling(f):
    w, h = 860, 640
    short = ["Used paper collected", "Sorted by type", "Shredded and mixed with water (pulp)", "Screened to remove plastic and staples",
             "Ink removed (soap and air bubbles)", "Pulp spread on a mesh; water drains", "Pressed between rollers", "Dried by heated rollers",
             "Wound onto large reels of new paper"]
    out = [text(w / 2, 38, f["title"], 20, "middle", "bold")]
    bw, bh = 210, 92
    xs = [60, 325, 590]
    ys = [80, 270, 460]
    # snake order: row 0 left→right, row 1 right→left, row 2 left→right
    order = [(0, 0), (1, 0), (2, 0), (2, 1), (1, 1), (0, 1), (0, 2), (1, 2), (2, 2)]
    centres = []
    for i, ((c, r), label) in enumerate(zip(order, short)):
        x, y = xs[c], ys[r]
        out.append(rect(x, y, bw, bh, "#ffffff", INK, 1.6, 10))
        out.append(lines(x + bw / 2, y + bh / 2 - (len(wrap(label, 22)) - 1) * 9 + 5, wrap(label, 22), 15, "middle"))
        out.append(badge(x + 4, y + 4, i + 1))
        centres.append((x, y))
    for i in range(len(order) - 1):
        (c0, r0), (c1, r1) = order[i], order[i + 1]
        x0, y0 = centres[i]
        x1, y1 = centres[i + 1]
        if r0 == r1 and c1 > c0:
            out.append(arrow(f"M{x0 + bw + 6},{y0 + bh / 2} L{x1 - 8},{y1 + bh / 2}"))
        elif r0 == r1:
            out.append(arrow(f"M{x0 - 6},{y0 + bh / 2} L{x1 + bw + 8},{y1 + bh / 2}"))
        else:
            out.append(arrow(f"M{x0 + bw / 2},{y0 + bh + 6} L{x1 + bw / 2},{y1 - 8}"))
    return svg(w, h, "\n".join(out), ARROW_DEFS)


def figures(tasks):
    return {
        "test1-task1.svg": population_tables(tasks["test1"]["facts"]),
        "test2-task1.svg": svg(740, 850, "\n".join(
            [text(370, 34, tasks["test2"]["facts"]["title"], 20, "middle", "bold")]
            + forest_map(30, 90, "1990", today=False) + forest_map(30, 490, "Today", today=True))),
        "test3-task1.svg": museum(tasks["test3"]["facts"]),
        "test4-task1.svg": recycling(tasks["test4"]["facts"]),
    }
