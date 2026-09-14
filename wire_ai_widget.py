#!/usr/bin/env python3
"""
Wire the "Analyze with TraceQube AI" widget into a tool page template.

Usage:
  python3 wire_ai_widget.py <file.html> --tool <tool_key> --target "<JS expr>" [--dry-run]

Examples:
  python3 wire_ai_widget.py web/templates/ping.html --tool ping --target host
  python3 wire_ai_widget.py web/templates/port-checker.html --tool port-checker \
      --target "document.getElementById('host').value.trim() + ':' + document.getElementById('port').value"
  python3 wire_ai_widget.py web/templates/what-is-my-ip.html --tool ip --target "'self'"

What it does:
  1. Finds <pre class="tq-result" id="X"></pre> and inserts the widget HTML right after it.
  2. Inserts the shared <style> block before the first <script>.
  3. Inserts the tqInitAIAnalyze() function at the top of that <script> block.
  4. Finds `const VAR = document.getElementById('X')` and the LAST `VAR.textContent = ...;`
     line, and appends a tqInitAIAnalyze({...}) call right after it.
  5. If found, hides the AI button/card when a new run starts
     (right after `VAR.style.display = 'block';`).

If any anchor isn't found, it aborts with an error and changes nothing —
safe to run repeatedly; already-wired files are skipped.
"""
import re, sys, argparse

WIDGET_HTML = r'''
  <div class="ai-widget-wrap">
    <button class="ai-analyze-btn" id="ai-analyze-btn" style="display:none">
      <svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 3v4M12 17v4M3 12h4M17 12h4M5.6 5.6l2.8 2.8M15.6 15.6l2.8 2.8M18.4 5.6l-2.8 2.8M8.4 15.6l-2.8 2.8"/></svg>
      Analyze with TraceQube AI
    </button>
    <div class="ai-result-card tq-card" id="ai-result-card" style="display:none"></div>
  </div>
'''

WIDGET_CSS = r'''
<style>
.ai-widget-wrap{margin:14px 0}
.ai-analyze-btn{display:inline-flex;align-items:center;gap:8px;background:linear-gradient(135deg,#6366f1,#8b5cf6);color:#fff;border:none;padding:10px 18px;border-radius:8px;font-size:14px;font-weight:600;cursor:pointer;transition:opacity .15s}
.ai-analyze-btn:hover{opacity:.9}
.ai-analyze-btn:disabled{opacity:.6;cursor:default}
.ai-analyze-btn .ai-spinner{width:14px;height:14px;border:2px solid rgba(255,255,255,.4);border-top-color:#fff;border-radius:50%;display:inline-block;animation:ai-spin .6s linear infinite}
@keyframes ai-spin{to{transform:rotate(360deg)}}
.ai-result-card{margin-top:12px;border-radius:8px;padding:16px}
.ai-result-header{font-weight:700;font-size:13px;text-transform:uppercase;letter-spacing:.05em;color:#6366f1;margin-bottom:8px}
.ai-result-body{font-size:14px;line-height:1.6}
.ai-remaining{margin-top:10px;font-size:12px;opacity:.6}
.ai-error{font-size:14px;color:#dc2626}
.ai-loading{font-size:14px;opacity:.6}
</style>
'''

WIDGET_JS_FUNC = r'''
function tqInitAIAnalyze(config) {
  const { tool, target, getOutput } = config;
  const btn = document.getElementById('ai-analyze-btn');
  const card = document.getElementById('ai-result-card');
  if (!btn || !card) return;
  card.style.display = 'none';
  card.innerHTML = '';
  btn.style.display = 'inline-flex';
  btn.disabled = false;
  btn.innerHTML = '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 3v4M12 17v4M3 12h4M17 12h4M5.6 5.6l2.8 2.8M15.6 15.6l2.8 2.8M18.4 5.6l-2.8 2.8M8.4 15.6l-2.8 2.8"/></svg> Analyze with TraceQube AI';
  btn.onclick = async function() {
    const output = getOutput();
    if (!output) return;
    btn.disabled = true;
    btn.innerHTML = '<span class="ai-spinner"></span> Analyzing...';
    card.style.display = 'block';
    card.innerHTML = '<p class="ai-loading">TraceQube AI is analyzing your results...</p>';
    try {
      const res = await fetch('/api/ai/analyze', {
        method: 'POST',
        headers: {'Content-Type': 'application/json'},
        body: JSON.stringify({ tool: tool, target: target, output: output })
      });
      const data = await res.json();
      if (!res.ok || data.error) {
        card.innerHTML = '<p class="ai-error">' + (data.error || 'Analysis unavailable. Please try again later.') + '</p>';
      } else {
        const left = data.remaining;
        const remainingText = (left > 0)
          ? left + ' free ' + (left === 1 ? 'analysis' : 'analyses') + ' remaining today'
          : 'Daily free limit reached \u2014 resets tomorrow';
        card.innerHTML =
          '<div class="ai-result-header">TraceQube AI Analysis</div>' +
          '<div class="ai-result-body">' + String(data.analysis).replace(/&/g,'&amp;').replace(/</g,'&lt;').replace(/\n/g,'<br>') + '</div>' +
          '<div class="ai-remaining">' + remainingText + '</div>';
      }
    } catch (e) {
      card.innerHTML = '<p class="ai-error">Could not reach analysis service. Please try again.</p>';
    } finally {
      btn.disabled = false;
      btn.innerHTML = '<svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 3v4M12 17v4M3 12h4M17 12h4M5.6 5.6l2.8 2.8M15.6 15.6l2.8 2.8M18.4 5.6l-2.8 2.8M8.4 15.6l-2.8 2.8"/></svg> Analyze with TraceQube AI';
    }
  };
}
'''

def main():
    p = argparse.ArgumentParser()
    p.add_argument('file')
    p.add_argument('--tool', required=True)
    p.add_argument('--target', required=True,
                    help="JS expression, e.g. 'host' or \"document.getElementById('ip').value.trim()\" or \"'self'\"")
    p.add_argument('--dry-run', action='store_true')
    args = p.parse_args()

    src = open(args.file).read()

    if 'ai-analyze-btn' in src:
        print(f"SKIP: {args.file} already wired")
        return

    # 1. Find result <pre> and its id
    m = re.search(r'(<pre class="tq-result" id="(\w+)"[^>]*></pre>)', src)
    if not m:
        sys.exit(f"ERROR: no <pre class=\"tq-result\" id=\"...\"> found in {args.file}")
    result_id = m.group(2)
    src = src[:m.end()] + "\n" + WIDGET_HTML + src[m.end():]

    # 2. Insert CSS before first <script>, JS func right after it opens
    m2 = re.search(r'<script>', src)
    if not m2:
        sys.exit(f"ERROR: no <script> tag found in {args.file}")
    src = src[:m2.start()] + WIDGET_CSS + "\n" + src[m2.start():]
    m2b = re.search(r'<script>', src)
    src = src[:m2b.end()] + WIDGET_JS_FUNC + src[m2b.end():]

    # 3. Find the JS variable bound to the result element
    m3 = re.search(r"(?:const|let|var)\s+(\w+)\s*=\s*document\.getElementById\('" + result_id + r"'\)", src)
    if not m3:
        sys.exit(f"ERROR: no variable assigned from document.getElementById('{result_id}') in {args.file}")
    rv = m3.group(1)

    # 4. Insert init call after the LAST `<rv>.textContent = ...;` line
    pattern = re.compile(re.escape(rv) + r"\.textContent\s*=.*?;\n")
    matches = list(pattern.finditer(src))
    if not matches:
        sys.exit(f"ERROR: no '{rv}.textContent = ...;' assignment found in {args.file}")
    last = matches[-1]
    init_call = f"  tqInitAIAnalyze({{ tool: '{args.tool}', target: {args.target}, getOutput: () => {rv}.textContent }});\n"
    src = src[:last.end()] + init_call + src[last.end():]

    # 5. Best-effort: hide widget while a new run is in progress
    m5 = re.search(re.escape(rv) + r"\.style\.display\s*=\s*'block';", src)
    if m5:
        reset = " document.getElementById('ai-analyze-btn').style.display='none'; document.getElementById('ai-result-card').style.display='none';"
        src = src[:m5.end()] + reset + src[m5.end():]
    else:
        print(f"WARN: couldn't find '{rv}.style.display = \\'block\\';' — skipped reset step (non-fatal)")

    if args.dry_run:
        print(src)
    else:
        open(args.file, 'w').write(src)
        print(f"OK: wired {args.file} (result var: {rv}, result id: {result_id})")

if __name__ == '__main__':
    main()
