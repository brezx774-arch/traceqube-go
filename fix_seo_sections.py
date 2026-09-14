#!/usr/bin/env python3
import re
import shutil
import sys
import os

FILES = [
    "dns-lookup.html",
    "down-checker.html",
    "http-header-checker.html",
    "ip-geolocation.html",
    "port-checker.html",
    "reverse-dns.html",
    "speed-test.html",
    "ssl-checker.html",
    "subnet-calculator.html",
    "whois-lookup.html",
]

TEMPLATE_DIR = "web/templates"
BACKUP_DIR = "web/templates/_backup_seo_fix"

# Pattern: closing div (seo-section #1) + closing div (container) + opening div (seo-section #2)
MERGE_PATTERN = re.compile(
    r'[ \t]*</div>\s*\n</div>\s*\n[ \t]*<div class="seo-section">\s*\n'
)

# Pattern: the ld+json script tag opening, to insert the missing container-close before it
SCRIPT_PATTERN = re.compile(r'([ \t]*<script type="application/ld\+json">)')

def fix_file(path):
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()

    original = content

    # Step 1: merge the two seo-sections into one (remove the premature container close)
    new_content, n1 = MERGE_PATTERN.subn("\n", content, count=1)
    if n1 != 1:
        print(f"  WARNING: merge pattern not found (or found {n1} times) in {path} — skipping")
        return False

    # Step 2: insert the missing closing </div> for .container right before the ld+json script
    new_content, n2 = SCRIPT_PATTERN.subn(r"  </div>\n\1", new_content, count=1)
    if n2 != 1:
        print(f"  WARNING: script tag pattern not found (or found {n2} times) in {path} — skipping")
        return False

    if new_content == original:
        print(f"  No changes made to {path}")
        return False

    with open(path, "w", encoding="utf-8") as f:
        f.write(new_content)

    return True

def main():
    os.makedirs(BACKUP_DIR, exist_ok=True)
    fixed = []
    for fname in FILES:
        path = os.path.join(TEMPLATE_DIR, fname)
        if not os.path.exists(path):
            print(f"SKIP (not found): {path}")
            continue
        backup_path = os.path.join(BACKUP_DIR, fname)
        shutil.copy2(path, backup_path)
        print(f"Backed up {fname} -> {backup_path}")
        ok = fix_file(path)
        if ok:
            fixed.append(fname)
            print(f"  FIXED: {fname}")

    print("\n=== Summary ===")
    print(f"Fixed {len(fixed)}/{len(FILES)} files: {', '.join(fixed)}")
    not_fixed = set(FILES) - set(fixed)
    if not_fixed:
        print(f"NOT fixed (check manually): {', '.join(not_fixed)}")

if __name__ == "__main__":
    main()
