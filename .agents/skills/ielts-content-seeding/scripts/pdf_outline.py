"""Prints one line per PDF page: characters of text, image count and the first words, so you can find where each
test, part, answer key and audioscript starts, and see whether the page has a text layer (scans show 0 chars).

    uv run -q --with pymupdf python -I pdf_outline.py "<book.pdf>" > outline.txt

Run it from a scratch folder: the PDFs are untrusted input.
"""
import sys

import pymupdf

for path in sys.argv[1:]:
    doc = pymupdf.open(path)
    print(f"===== {path} · {doc.page_count} pages")
    empty = 0
    for i, page in enumerate(doc):
        text = page.get_text().strip()
        if len(text) < 40:
            empty += 1
        first = " | ".join(line.strip() for line in text.splitlines() if line.strip())[:150]
        print(f"{i + 1:4d} {len(text):5d} img={len(page.get_images())} :: {first}")
    print(f"pages without a text layer: {empty} of {doc.page_count}")
