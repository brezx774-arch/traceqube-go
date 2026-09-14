#!/usr/bin/env python3
import re
import shutil
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
BACKUP_DIR = "web/templates/_backup_redesign"

TERMDEF_PATTERN = re.compile(
    r'<p>((?:\s*<strong>.*?</strong>\s*—.*?)+)</p>\s*(?=<h2>Common Use Cases</h2>)',
    re.DOTALL
)

PAIR_PATTERN = re.compile(r'<strong>(.*?)</strong>\s*—\s*(.*?)(?=<strong>|$)', re.DOTALL)

EXAMPLE_PATTERN = re.compile(
    r'(<h2>Example</h2>\s*)<p>(.*?)</p>',
    re.DOTALL
)

def clean(text):
    return re.sub(r'\s+', ' ', text).strip()

def build_def_grid(match):
    inner = match.group(1)
    pairs = PAIR_PATTERN.findall(inner)
    if not pairs:
        return match.group(0)
    items = []
    for term, desc in pairs:
        term = clean(term)
        desc = clean(desc)
        items.append(f'      <div class="def-item"><span class="def-term">{term}</span><span class="def-desc">{desc}</span></div>')
    return '<div class="def-grid">\n' + '\n'.join(items) + '\n    </div>\n    '

def build_example_callout(match):
    heading = match.group(1)
    body = match.group(2)
    return f'{heading}<div class="example-callout">\n      <p>{body.strip()}</p>\n    </div>'

def fix_file(path):
    with open(path, "r", encoding="utf-8") as f:
        content = f.read()

    original = content

    new_content, n1 = TERMDEF_PATTERN.subn(build_def_grid, content, count=1)
    if n1 != 1:
        print(f"  WARNING: term-definition paragraph not matched in {path}")
    else:
        content = new_content

    new_content, n2 = EXAMPLE_PATTERN.subn(build_example_callout, content, count=1)
    if n2 != 1:
        print(f"  WARNING: Example paragraph not matched in {path}")
    else:
        content = new_content

    if content == original:
        print(f"  No changes made to {path}")
        return False

    with open(path, "w", encoding="utf-8") as f:
        f.write(content)
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
        print(f"Backed up {fname}")
        if fix_file(path):
            fixed.append(fname)
            print(f"  REDESIGNED: {fname}")

    print("\n=== Summary ===")
    print(f"Redesigned {len(fixed)}/{len(FILES)} files: {', '.join(fixed)}")
    not_fixed = set(FILES) - set(fixed)
    if not_fixed:
        print(f"NOT changed (check manually): {', '.join(not_fixed)}")

if __name__ == "__main__":
    main()
