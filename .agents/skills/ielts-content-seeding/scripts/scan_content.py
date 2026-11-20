"""Quick content scan of a set: word counts, question types, answer mix, gap answers missing from the text,
explanations without paraphrase pairs, and any mention of the reference book.

    python3 .agents/skills/ielts-content-seeding/scripts/scan_content.py be/seeds/exam/s1
"""
import collections
import glob
import json
import re
import sys

GAP_TYPES = {"notes", "sentence", "summary", "table", "flowchart", "form", "short_answer", "diagram"}

for path in sorted(glob.glob(f"{sys.argv[1].rstrip('/')}/test*/*.json")):
    d = json.load(open(path))
    text = " ".join(p["text"] for p in d.get("paragraphs", [])) or " ".join(l["text"] for l in d.get("lines", []))
    key = d["key"]
    issues = []
    for g in d["groups"]:
        if g["type"] not in GAP_TYPES:
            continue
        for n in range(g["from"], g["to"] + 1):
            # Any accepted form may be the one in the text ("8" is spoken as "eight").
            forms = [re.sub(r"\(.*?\)", "", alt).strip() for a in key[str(n)]["answer"] for alt in a.split("/")]
            # Listening numbers and dates are spoken as words ("twelfth of April"); check them by ear instead.
            if d.get("lines") and any(ch.isdigit() for f in forms for ch in f):
                continue
            if not any(f and f.lower().lstrip("£$") in text.lower() for f in forms):
                issues.append(f"Q{n} answer {forms[0]!r} not in text")
    no_pairs = [k for k, v in key.items() if not v.get("pairs") and v["answer"][0] != "NOT GIVEN"]
    if no_pairs:
        issues.append(f"no pairs: {no_pairs}")
    if "cambridge" in open(path).read().lower():
        issues.append("mentions Cambridge")
    mix = collections.Counter(v["answer"][0] for v in key.values() if len(v["answer"][0]) <= 9)
    types = " ".join(f'{g["type"]}×{g["to"] - g["from"] + 1}' for g in d["groups"])
    audio = f' · {d["audio"]["duration"]:.0f}s audio' if d.get("audio") else ""
    print(f"{path}: {d['title'][:48]!r} · {len(text.split())} words{audio} · {types}")
    print(f"    answers {dict(mix.most_common(8))} · {'; '.join(issues) or 'ok'}")
