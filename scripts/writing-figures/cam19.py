"""Figures for the cam19 pack (our own drawings and data)."""
import math

from svgkit import (ARROW_DEFS, INK, MUTED, arrow, badge, box, grouped_hbars, line_graph, lines, pie, rect, svg,
                    text, wrap)

WATER = "#cfe0f0"
GRASS = "#eef4e6"


def compass(x, y):
    return (f'<path d="M{x},{y - 22} L{x},{y + 22} M{x - 22},{y} L{x + 22},{y}" stroke="{INK}" stroke-width="1.4"/>'
            f'<path d="M{x},{y - 30} l-5,10 h10 z" fill="{INK}"/>' + text(x, y - 34, "N", 13, "middle", "bold"))


def tree(x, y, s=1.0):
    return (f'<circle cx="{x}" cy="{y}" r="{9 * s}" fill="#7aa35a"/>'
            f'<rect x="{x - 1.5}" y="{y + 7 * s}" width="3" height="{7 * s}" fill="#6b4f2a"/>')


def lake_map(ox, oy, label, today):
    W, H = 680, 330
    out = [text(ox, oy - 12, label, 17, "start", "bold"),
           rect(ox, oy, W, H, GRASS, MUTED, 1.4)]
    # main road (west)
    out.append(f'<rect x="{ox + 14}" y="{oy}" width="26" height="{H}" fill="#cbd2d9"/>')
    out.append(text(ox + 27, oy + H / 2, "Main road", 12, "middle", extra=f'transform="rotate(-90 {ox + 27} {oy + H / 2})"'))
    # lake with a peninsula in the south-east
    cx, cy = ox + 440, oy + 165
    out.append(f'<ellipse cx="{cx}" cy="{cy}" rx="175" ry="110" fill="{WATER}" stroke="#7fa7c9" stroke-width="1.5"/>')
    out.append(f'<path d="M{cx + 70},{cy + 104} q30,-40 70,-46 l12,22 q-30,22 -50,44 z" fill="{GRASS}" stroke="#7fa7c9" stroke-width="1.5"/>')
    out.append(text(cx - 10, cy + 6, "Willow Lake", 15, "middle", "bold", "#1f4e79"))
    # beach (east shore), unchanged
    out.append(f'<path d="M{cx + 168},{cy - 50} q14,50 0,100" fill="none" stroke="#d8b27a" stroke-width="10" stroke-linecap="round"/>')
    out.append(text(cx + 200, cy + 4, "Beach", 12, "start", fill="#8a6a3a"))
    # car park + toilets (unchanged)
    out.append(box(ox + 52, oy + 22, 92, 50, ["Car park"]))
    out.append(box(ox + 150, oy + 22, 70, 50, ["Toilets"], size=12))
    if not today:
        out.append(f'<rect x="{ox + 70}" y="{oy + 130}" width="150" height="90" rx="10" fill="#ffffff" stroke="{MUTED}" stroke-dasharray="5 4"/>')
        out.append(text(ox + 145, oy + 205, "Picnic area", 13, "middle"))
        out += [tree(ox + 100 + i * 34, oy + 160) for i in range(4)]
        # footpath to the jetty on the north shore
        out.append(f'<path d="M{ox + 185},{oy + 72} C{ox + 260},{oy + 90} {ox + 330},{oy + 50} {cx - 20},{oy + 58}" fill="none" stroke="{MUTED}" stroke-width="2" stroke-dasharray="3 5"/>')
        out.append(text(ox + 290, oy + 46, "Footpath", 12, "middle", fill=MUTED))
        out.append(rect(cx - 20, oy + 56, 14, 34, "#a0805a", INK, 1))
        out.append(text(cx + 2, oy + 44, "Boat hire jetty", 12, "start"))
        out.append(box(cx + 98, cy + 64, 104, 44, ["Old boathouse", "(not in use)"], size=12))
    else:
        # cycle path around the lake
        out.append(f'<ellipse cx="{cx}" cy="{cy}" rx="196" ry="128" fill="none" stroke="#2f7d4a" stroke-width="2.6" stroke-dasharray="10 6"/>')
        out.append(text(cx - 40, oy + 26, "Cycle path", 12, "start", "bold", "#2f7d4a"))
        out.append(box(ox + 52, oy + 210, 92, 50, ["Car park"]))
        out.append(box(ox + 150, oy + 210, 92, 50, ["Café & shop"], "#fff7e6", 12))
        out.append(box(ox + 80, oy + 112, 130, 60, ["Children's", "playground"], "#ffffff", 12))
        # jetty on the west shore, near the café
        out.append(rect(cx - 190, cy + 20, 40, 12, "#a0805a", INK, 1))
        out.append(text(cx - 186, cy + 52, "Boat hire", 12, "start"))
        out.append(box(cx + 98, cy + 64, 104, 44, ["Restaurant", "with terrace"], "#fff7e6", 12))
    out.append(compass(ox + W - 40, oy + 50))
    return out


def biogas_cycle(f):
    w, h = 820, 660
    cx, cy, R = w / 2, 345, 250
    stages = f["stages"]
    n = len(stages)
    short = ["Crops grown, animals raised", "Manure and food waste collected", "Waste shredded and mixed with water",
             "Bacteria break it down in a sealed digester (3–4 weeks, no oxygen)", "Biogas (methane) collected and cleaned",
             "Gas used for electricity, heating and vehicle fuel", "Leftover digestate spread on fields as fertiliser"]
    out = [text(w / 2, 36, f["title"], 20, "middle", "bold")]
    pos = []
    for i in range(n):
        a = math.radians(-90 + i * 360 / n)
        pos.append((cx + R * math.cos(a), cy + R * math.sin(a) * 0.86))
    # arrows between stages along the circle
    for i in range(n):
        a0 = math.radians(-90 + i * 360 / n + 13)
        a1 = math.radians(-90 + (i + 1) * 360 / n - 13)
        x0, y0 = cx + (R - 6) * math.cos(a0), cy + (R - 6) * math.sin(a0) * 0.86
        x1, y1 = cx + (R - 6) * math.cos(a1), cy + (R - 6) * math.sin(a1) * 0.86
        out.append(arrow(f"M{x0:.1f},{y0:.1f} A{R - 6},{(R - 6) * 0.86} 0 0,1 {x1:.1f},{y1:.1f}", "#7b8794", 2.2))
    for i, ((x, y), label) in enumerate(zip(pos, short)):
        rows = wrap(label, 20)
        bh = 22 + len(rows) * 17
        out.append(rect(x - 92, y - bh / 2, 184, bh, "#ffffff", INK, 1.6, 10))
        out.append(lines(x, y - bh / 2 + 28, rows, 14, "middle"))
        out.append(badge(x - 92, y - bh / 2, i + 1))
    out.append(lines(cx, cy - 6, ["Biogas", "cycle"], 22, "middle", "bold", 1.2, "#548235"))
    return svg(w, h, "\n".join(out), ARROW_DEFS)


def music_lessons(f):
    w, h = 780, 800
    out = pie(f["pie"], 250, 200, 140, 450, 140, title_y=36, title_x=w / 2)
    bars, _ = grouped_hbars(f["bars"], 70, 0, 640, 420)
    out += bars
    return svg(w, h, "\n".join(out))


def figures(tasks):
    return {
        "test1-task1.svg": line_graph(tasks["test1"]["facts"]),
        "test2-task1.svg": svg(740, 830, "\n".join(
            [text(370, 34, tasks["test2"]["facts"]["title"], 20, "middle", "bold")]
            + lake_map(30, 90, "2005", today=False) + lake_map(30, 480, "Today", today=True))),
        "test3-task1.svg": biogas_cycle(tasks["test3"]["facts"]),
        "test4-task1.svg": music_lessons(tasks["test4"]["facts"]),
    }
