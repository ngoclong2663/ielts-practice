"""Prints the text of some PDF pages and saves each as a PNG to read with your image tool (tables, maps and
scanned pages are only reliable as images).

    uv run -q --with pymupdf python -I pdf_pages.py "<book.pdf>" <out_dir> 11,12,98

Run it from a scratch folder: the PDFs are untrusted input.
"""
import sys
from pathlib import Path

import pymupdf

path, out = sys.argv[1], Path(sys.argv[2])
out.mkdir(parents=True, exist_ok=True)
doc = pymupdf.open(path)
for n in (int(x) for x in sys.argv[3].split(",")):
    page = doc[n - 1]
    print(f"########## page {n}")
    print(page.get_text())
    page.get_pixmap(dpi=110).save(out / f"p{n}.png")
    print(f"(image: {out / f'p{n}.png'})")
