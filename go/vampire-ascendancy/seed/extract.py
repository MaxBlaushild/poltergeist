#!/usr/bin/env python3
"""Extract the Vampire Ascendancy character packets from the master PDF into
structured JSON (seed/characters.json).

The master doc is prose-heavy and still being revised, so this step is meant to
be re-run whenever the doc changes. The emitted characters.json is the editable
source of truth the Go seed importer loads.

Usage:
    python3 extract.py "/path/to/Vampire Character Packets Master File.pdf"

Requires: pypdf  (pip install --user pypdf)
"""
import json
import re
import sys
import unicodedata
from pathlib import Path

HOUSES = ["Spires", "Chains", "Cinders", "Ashglass", "Marquess's Court"]
HOUSE_SORT = {name: i for i, name in enumerate(HOUSES)}

SECTION_LABELS = ["PRE-EVENT INFO", "POST-ACT 1 CONTEXT", "SECRETS", "MISSIONS"]


def norm(s: str) -> str:
    # Normalize unicode (curly quotes -> straight) and collapse runs of spaces.
    s = unicodedata.normalize("NFKC", s)
    s = s.replace("’", "'").replace("‘", "'")
    s = s.replace("“", '"').replace("”", '"')
    s = re.sub(r"[ \t]+", " ", s)
    return s


def detect_house(rest: str):
    """Return (house_name, rest_without_house)."""
    r = rest
    if "Marquess's Court" in r:
        return "Marquess's Court", r.replace("Marquess's Court", " ")
    m = re.search(r"House of (Spires|Chains|Cinders|Ashglass)", r)
    if m:
        return m.group(1), r[: m.start()] + " " + r[m.end():]
    # NPC note like "Ashglass (order of Father Mirth)"
    m = re.search(r"\b(Spires|Chains|Cinders|Ashglass)\b", r)
    if m:
        return m.group(1), r[: m.start()] + " " + r[m.end():]
    return None, r


def classify_role(rest: str) -> str:
    up = rest.upper()
    if "NOTE" in up or re.search(r"\bNPC\b", up) and "GM/NPC" not in up:
        # Explicit NPC reference notes.
        if "NOTE" in up:
            return "npc"
    if "GM" in up:  # ★ GM or GM/NPC -> GM-run packet
        return "gm"
    if re.search(r"\bNPC\b", up):
        return "npc"
    return "player"


def clean_title(rest: str) -> str:
    t = rest
    # Strip marker tokens.
    for tok in ["GM/NPC", "(GM/NPC)", "OPTIONAL", "NPC", "NOTE", "GM"]:
        t = t.replace(tok, " ")
    t = t.replace("★", " ").replace("✦", " ")  # ★ ✦
    t = t.replace("(", " ").replace(")", " ")
    # Collapse leftover separators/whitespace and trim dangling em dashes.
    t = re.sub(r"[ \t]+", " ", t).strip()
    t = t.strip("—-– ").strip()
    # Trim a trailing/leading em dash with spaces in the middle is fine to keep.
    return t


def parse_secrets(text: str):
    text = text.strip()
    if not text:
        return []
    # Split on "1." "2." "3." markers.
    parts = re.split(r"(?:^|\s)(\d+)\.\s+", text)
    secrets = []
    # parts = ['', '1', 'body1', '2', 'body2', ...] when leading marker present
    it = iter(parts)
    first = next(it, "")
    if first.strip():
        # No leading number; treat whole thing as one secret.
        secrets.append(first.strip())
    pending_ordinal = None
    for chunk in it:
        if pending_ordinal is None:
            pending_ordinal = chunk
        else:
            secrets.append(chunk.strip())
            pending_ordinal = None
    return [s for s in secrets if s]


TIER_BT = {"easy": 2, "medium": 4, "hard": 6}


def parse_missions(text: str):
    text = text.strip()
    if not text or text.strip(" -—") == "":
        return []
    blocks = re.split(r"Mission\s+\d+\s*[—\-]\s*", text)
    missions = []
    ordinal = 0
    for b in blocks:
        b = b.strip()
        if not b:
            continue
        ordinal += 1
        tier_m = re.match(r"(Easy|Medium|Hard)", b, re.IGNORECASE)
        tier = tier_m.group(1).lower() if tier_m else "easy"
        bt_m = re.search(r"\((\d+)\s*BT\)", b)
        reward = int(bt_m.group(1)) if bt_m else TIER_BT.get(tier, 0)
        # Remove the "Tier (n BT)" prefix to get the body.
        body = b
        if bt_m:
            body = b[bt_m.end():].strip()
        elif tier_m:
            body = b[tier_m.end():].strip()
        # Split off a trailing "Enter ..." answer instruction as answer_format.
        answer_format = ""
        em = re.search(r"(Enter\b.*)$", body, re.DOTALL)
        if em:
            answer_format = em.group(1).strip()
            body = body[: em.start()].strip()
        missions.append(
            {
                "ordinal": ordinal,
                "tier": tier,
                "reward_bt": reward,
                "prompt": body,
                "answer_format": answer_format,
            }
        )
    return missions


def slice_sections(block: str):
    """Return dict label -> text for the section labels present in block."""
    positions = []
    for label in SECTION_LABELS:
        idx = block.find(label)
        if idx != -1:
            positions.append((idx, label))
    positions.sort()
    out = {}
    for i, (idx, label) in enumerate(positions):
        start = idx + len(label)
        end = positions[i + 1][0] if i + 1 < len(positions) else len(block)
        out[label] = block[start:end].strip()
    return out


def main():
    if len(sys.argv) < 2:
        print("usage: extract.py <path-to-pdf> [out.json]", file=sys.stderr)
        sys.exit(1)
    pdf_path = sys.argv[1]
    out_path = sys.argv[2] if len(sys.argv) > 2 else str(
        Path(__file__).parent / "characters.json"
    )

    import pypdf

    reader = pypdf.PdfReader(pdf_path)
    text = norm("\n".join((p.extract_text() or "") for p in reader.pages))

    lines = text.split("\n")
    # A header line looks like "<Name> | <title...> <house>"
    header_idxs = [i for i, l in enumerate(lines) if " | " in l]

    characters = []
    for hi, line_idx in enumerate(header_idxs):
        header = lines[line_idx].strip()
        name, rest = header.split(" | ", 1)
        name = name.strip()
        house, rest_no_house = detect_house(rest)
        role = classify_role(rest)
        is_optional = "OPTIONAL" in rest.upper() or "✦" in rest
        title = clean_title(rest_no_house)

        # Body is everything from after this header line up to the next header line.
        end_line = header_idxs[hi + 1] if hi + 1 < len(header_idxs) else len(lines)
        block = "\n".join(lines[line_idx + 1 : end_line])
        sections = slice_sections(block)

        characters.append(
            {
                "name": name,
                "title": title,
                "house": house,
                "role_type": role,
                "is_optional": is_optional,
                "pre_event_info": sections.get("PRE-EVENT INFO", "").strip(),
                "post_act1_context": sections.get("POST-ACT 1 CONTEXT", "").strip(),
                "secrets": parse_secrets(sections.get("SECRETS", "")),
                "missions": parse_missions(sections.get("MISSIONS", "")),
            }
        )

    payload = {
        "houses": [{"name": h, "sort_order": HOUSE_SORT[h]} for h in HOUSES],
        "characters": characters,
    }
    with open(out_path, "w") as f:
        json.dump(payload, f, indent=2, ensure_ascii=False)

    # Summary to stderr for a quick sanity check.
    print(f"wrote {out_path}", file=sys.stderr)
    print(f"characters: {len(characters)}", file=sys.stderr)
    by_role = {}
    for c in characters:
        by_role[c["role_type"]] = by_role.get(c["role_type"], 0) + 1
    print(f"by role: {by_role}", file=sys.stderr)
    missing = [c["name"] for c in characters if not c["house"]]
    if missing:
        print(f"WARNING no house: {missing}", file=sys.stderr)


if __name__ == "__main__":
    main()
