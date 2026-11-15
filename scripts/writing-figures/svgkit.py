"""Small SVG toolkit for the writing task figures (Epic 5 §6).

Charts are drawn from a task's `facts` in pack.json, so the figure and the data the AI sees always match.
Maps and process diagrams are drawn per pack with the shape helpers below. Output is plain SVG: white
background, system sans-serif, no external resources (the API serves it from S3 or the binary).
"""
import math
from xml.sax.saxutils import escape

FONT = "Arial, Helvetica, sans-serif"
INK = "#1f2933"
MUTED = "#52606d"
GRID = "#d9dee3"
SERIES = [  # colour, dash pattern, marker
    ("#1f4e79", "", "circle"),
    ("#c55a11", "8 5", "triangle"),
    ("#548235", "2 4", "square"),
    ("#7030a0", "12 4 2 4", "diamond"),
    ("#b8860b", "4 3", "cross"),
]
FILLS = ["#1f4e79", "#5b8cc0", "#c55a11", "#e8a87c", "#7b8794", "#548235"]
LIGHT_FILLS = {"#e8a87c"}


def svg(w, h, body, defs=""):
    return (
        f'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 {w} {h}" width="{w}" height="{h}" '
        f'font-family="{FONT}" fill="{INK}">\n'
        + (f"<defs>{defs}</defs>\n" if defs else "")
        + f'<rect width="{w}" height="{h}" fill="#ffffff"/>\n{body}\n</svg>\n'
    )


ARROW_DEFS = (
    '<marker id="arrow" viewBox="0 0 10 10" refX="8" refY="5" markerWidth="7" markerHeight="7" orient="auto-start-reverse">'
    f'<path d="M0,0 L10,5 L0,10 z" fill="{INK}"/></marker>'
)


def text(x, y, s, size=15, anchor="start", weight="normal", fill=INK, extra=""):
    return (
        f'<text x="{x:.1f}" y="{y:.1f}" font-size="{size}" text-anchor="{anchor}" font-weight="{weight}" '
        f'fill="{fill}" {extra}>{escape(str(s))}</text>'
    )


def lines(x, y, rows, size=15, anchor="start", weight="normal", lh=1.25, fill=INK):
    return "\n".join(text(x, y + i * size * lh, r, size, anchor, weight, fill) for i, r in enumerate(rows))


def wrap(s, width):
    """Greedy word wrap by character count."""
    out, cur = [], ""
    for word in s.split():
        if cur and len(cur) + 1 + len(word) > width:
            out.append(cur)
            cur = word
        else:
            cur = f"{cur} {word}".strip()
    if cur:
        out.append(cur)
    return out


def rect(x, y, w, h, fill="#ffffff", stroke=INK, sw=1.6, rx=0, dash=""):
    d = f' stroke-dasharray="{dash}"' if dash else ""
    return f'<rect x="{x:.1f}" y="{y:.1f}" width="{w:.1f}" height="{h:.1f}" rx="{rx}" fill="{fill}" stroke="{stroke}" stroke-width="{sw}"{d}/>'


def box(x, y, w, h, label_lines, fill="#ffffff", size=13, stroke=INK, dash="", weight="normal"):
    cy = y + h / 2 - (len(label_lines) - 1) * size * 0.62 + size * 0.35
    return rect(x, y, w, h, fill, stroke, 1.6, 0, dash) + "\n" + lines(x + w / 2, cy, label_lines, size, "middle", weight)


def arrow(path, color=INK, width=2.4):
    return f'<path d="{path}" fill="none" stroke="{color}" stroke-width="{width}" marker-end="url(#arrow)"/>'


def badge(x, y, n, r=13):
    return f'<circle cx="{x}" cy="{y}" r="{r}" fill="{INK}"/>' + text(x, y + 5, str(n), 14, "middle", "bold", "#ffffff")


def marker(kind, x, y, color):
    if kind == "circle":
        return f'<circle cx="{x:.1f}" cy="{y:.1f}" r="5.5" fill="#fff" stroke="{color}" stroke-width="2.2"/>'
    if kind == "square":
        return f'<rect x="{x - 5:.1f}" y="{y - 5:.1f}" width="10" height="10" fill="{color}"/>'
    if kind == "triangle":
        return f'<polygon points="{x:.1f},{y - 6.5:.1f} {x + 6:.1f},{y + 4.5:.1f} {x - 6:.1f},{y + 4.5:.1f}" fill="#fff" stroke="{color}" stroke-width="2"/>'
    if kind == "cross":
        return (f'<path d="M{x - 5:.1f},{y - 5:.1f} L{x + 5:.1f},{y + 5:.1f} M{x - 5:.1f},{y + 5:.1f} L{x + 5:.1f},{y - 5:.1f}" '
                f'stroke="{color}" stroke-width="2.4"/>')
    return f'<polygon points="{x:.1f},{y - 6.5:.1f} {x + 6.5:.1f},{y:.1f} {x:.1f},{y + 6.5:.1f} {x - 6.5:.1f},{y:.1f}" fill="{color}"/>'


def fmt(v):
    return f"{v:,}" if isinstance(v, int) else str(v)


# ---- charts ------------------------------------------------------------------------------------


def line_graph(f, w=780, h=500):
    """facts: title, x_label, y_label, y_max, y_step?, categories, series[{name, values}]"""
    left, right, top, bottom = 80, 560, 70, 420
    cats, ymax, step = f["categories"], f["y_max"], f.get("y_step", 2)
    xs = lambda i: left + i * (right - left) / (len(cats) - 1)
    ys = lambda v: bottom - v / ymax * (bottom - top)
    out = [text(w / 2, 36, f["title"], 19, "middle", "bold")]
    for v in range(0, ymax + 1, step):
        out.append(f'<line x1="{left}" y1="{ys(v):.1f}" x2="{right}" y2="{ys(v):.1f}" stroke="{GRID}"/>')
        out.append(text(left - 10, ys(v) + 5, fmt(v), 14, "end", fill=MUTED))
    out.append(f'<line x1="{left}" y1="{top}" x2="{left}" y2="{bottom}" stroke="{INK}" stroke-width="1.5"/>')
    out.append(f'<line x1="{left}" y1="{bottom}" x2="{right}" y2="{bottom}" stroke="{INK}" stroke-width="1.5"/>')
    for i, c in enumerate(cats):
        out.append(text(xs(i), bottom + 26, c, 14, "middle", fill=MUTED))
    out.append(text((left + right) / 2, bottom + 56, f["x_label"], 15, "middle", "bold"))
    cy = (top + bottom) / 2
    out.append(text(26, cy, f["y_label"], 15, "middle", "bold", extra=f'transform="rotate(-90 26 {cy})"'))
    for s, (color, dash, mk) in zip(f["series"], SERIES):
        pts = " ".join(f"{xs(i):.1f},{ys(v):.1f}" for i, v in enumerate(s["values"]))
        da = f' stroke-dasharray="{dash}"' if dash else ""
        out.append(f'<polyline points="{pts}" fill="none" stroke="{color}" stroke-width="2.6"{da}/>')
        out += [marker(mk, xs(i), ys(v), color) for i, v in enumerate(s["values"])]
    lx, ly = 594, 140
    out.append(rect(lx - 14, ly - 30, 192, len(f["series"]) * 34 + 20, stroke=GRID, sw=1))
    for i, (s, (color, dash, mk)) in enumerate(zip(f["series"], SERIES)):
        y = ly + i * 34
        da = f' stroke-dasharray="{dash}"' if dash else ""
        out.append(f'<line x1="{lx}" y1="{y - 5}" x2="{lx + 40}" y2="{y - 5}" stroke="{color}" stroke-width="2.6"{da}/>')
        out.append(marker(mk, lx + 20, y - 5, color))
        out.append(text(lx + 52, y, s["name"], 15))
    return svg(w, h, "\n".join(out))


def pie(f, cx, cy, r, legend_x, legend_y, title_y=None, title_x=None):
    """facts: title, slices[{label, value}] (percent). Returns SVG fragments."""
    out = []
    if title_y is not None:
        out.append(text(title_x if title_x is not None else cx, title_y, f["title"], 17, "middle", "bold"))
    ang = -90.0
    for s, color in zip(f["slices"], FILLS):
        sweep = s["value"] / 100 * 360
        a0, a1 = math.radians(ang), math.radians(ang + sweep)
        x0, y0 = cx + r * math.cos(a0), cy + r * math.sin(a0)
        x1, y1 = cx + r * math.cos(a1), cy + r * math.sin(a1)
        large = 1 if sweep > 180 else 0
        out.append(f'<path d="M{cx},{cy} L{x0:.1f},{y0:.1f} A{r},{r} 0 {large},1 {x1:.1f},{y1:.1f} z" fill="{color}" stroke="#fff" stroke-width="2"/>')
        mid = math.radians(ang + sweep / 2)
        lx, ly = cx + r * 0.66 * math.cos(mid), cy + r * 0.66 * math.sin(mid)
        out.append(text(lx, ly + 5, f'{s["value"]}%', 15, "middle", "bold", INK if color in LIGHT_FILLS else "#ffffff"))
        ang += sweep
    for i, (s, color) in enumerate(zip(f["slices"], FILLS)):
        y = legend_y + i * 32
        out.append(f'<rect x="{legend_x}" y="{y - 13}" width="16" height="16" fill="{color}"/>')
        out.append(text(legend_x + 26, y, s["label"], 15))
    return out


def grouped_hbars(f, x0, y0, width, title_y):
    """facts: title, x_label, x_max, x_step, groups[2], categories[{label, values[2]}]. Returns fragments."""
    out = [text(x0 + width / 2, title_y, f["title"], 17, "middle", "bold")]
    colors = ["#5b8cc0", "#c55a11"]
    for i, (g, c) in enumerate(zip(f["groups"], colors)):
        lx = x0 + width / 2 - 110 + i * 130
        out.append(f'<rect x="{lx}" y="{title_y + 14}" width="16" height="16" fill="{c}"/>')
        out.append(text(lx + 24, title_y + 27, g, 14))
    top, label_w, bar_h, gap = title_y + 52, 90, 20, 22
    plot_w = width - label_w
    sx = lambda v: x0 + label_w + v / f["x_max"] * plot_w
    rows = f["categories"]
    bottom = top + len(rows) * (2 * bar_h + gap)
    for v in range(0, f["x_max"] + 1, f["x_step"]):
        out.append(f'<line x1="{sx(v):.1f}" y1="{top - 6}" x2="{sx(v):.1f}" y2="{bottom}" stroke="{GRID}"/>')
        out.append(text(sx(v), bottom + 20, fmt(v), 13, "middle", fill=MUTED))
    out.append(f'<line x1="{sx(0)}" y1="{top - 6}" x2="{sx(0)}" y2="{bottom}" stroke="{INK}" stroke-width="1.5"/>')
    for i, row in enumerate(rows):
        y = top + i * (2 * bar_h + gap)
        out.append(text(x0 + label_w - 12, y + bar_h + 5, row["label"], 15, "end"))
        for j, (v, c) in enumerate(zip(row["values"], colors)):
            out.append(f'<rect x="{sx(0)}" y="{y + j * bar_h}" width="{sx(v) - sx(0):.1f}" height="{bar_h - 2}" fill="{c}"/>')
    out.append(text(x0 + label_w + plot_w / 2, bottom + 44, f["x_label"], 14, "middle", "bold"))
    return out, bottom + 50


def vbars(f, x0, y0, width, height):
    """facts: title, y_max, y_step, categories, values. Returns fragments."""
    out = [text(x0 + width / 2, y0, f["title"], 17, "middle", "bold")]
    left, top, bottom = x0 + 70, y0 + 24, y0 + height - 40
    right = x0 + width
    sy = lambda v: bottom - v / f["y_max"] * (bottom - top)
    for v in range(0, f["y_max"] + 1, f["y_step"]):
        out.append(f'<line x1="{left}" y1="{sy(v):.1f}" x2="{right}" y2="{sy(v):.1f}" stroke="{GRID}"/>')
        out.append(text(left - 8, sy(v) + 5, fmt(v), 13, "end", fill=MUTED))
    out.append(f'<line x1="{left}" y1="{top}" x2="{left}" y2="{bottom}" stroke="{INK}" stroke-width="1.5"/>')
    out.append(f'<line x1="{left}" y1="{bottom}" x2="{right}" y2="{bottom}" stroke="{INK}" stroke-width="1.5"/>')
    n = len(f["categories"])
    slot = (right - left) / n
    for i, (c, v) in enumerate(zip(f["categories"], f["values"])):
        bx = left + i * slot + slot * 0.2
        out.append(f'<rect x="{bx:.1f}" y="{sy(v):.1f}" width="{slot * 0.6:.1f}" height="{bottom - sy(v):.1f}" fill="#5b8cc0"/>')
        out.append(text(left + i * slot + slot / 2, bottom + 20, c, 13, "middle", fill=MUTED))
    return out


def table(title, columns, rows, x0, y0, col_w, row_h=36, title_size=17):
    """rows: list of lists of strings (first column is the row label). Returns (fragments, bottom y)."""
    out = [text(x0 + sum(col_w) / 2, y0, title, title_size, "middle", "bold")]
    y = y0 + 14
    for ri, row in enumerate([columns] + rows):
        x = x0
        head = ri == 0
        for ci, cell in enumerate(row):
            out.append(rect(x, y, col_w[ci], row_h, "#eef2f6" if head else "#ffffff", MUTED, 1))
            if ci == 0:
                out.append(text(x + 12, y + row_h / 2 + 5, cell, 14, "start", "bold" if head else "normal"))
            else:
                out.append(text(x + col_w[ci] / 2, y + row_h / 2 + 5, cell, 14, "middle", "bold" if head else "normal"))
            x += col_w[ci]
        y += row_h
    return out, y
