"""Draws the Task 1 figures of a writing seed pack.

Usage (from the repo root):  python3 scripts/writing-figures/build.py cam19
Reads be/seeds/writing/<pack>/pack.json and writes be/seeds/writing/<pack>/figures/*.svg.
Each pack has a module here (cam19.py, ...) with figures(tasks) -> {file name: svg}.
Then run `make seed` to upload changed figures and update the tasks.
"""
import importlib
import json
import sys
from pathlib import Path

ROOT = Path(__file__).resolve().parents[2]
sys.path.insert(0, str(Path(__file__).resolve().parent))

pack = sys.argv[1]
pack_dir = ROOT / "be" / "seeds" / "writing" / pack
data = json.loads((pack_dir / "pack.json").read_text())
tasks = {f"test{t['test_no']}": t for t in data["tasks"] if t["task_no"] == 1}
out = pack_dir / "figures"
out.mkdir(exist_ok=True)
for name, content in importlib.import_module(pack).figures(tasks).items():
    (out / name).write_text(content)
    print("wrote", out / name)
