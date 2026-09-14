import os, re, json

BASE = os.path.expanduser("~/traceqube-go/web/templates")

FAQS = {
"dns-lookup.html": [
    ("Why do I see different results than my own computer?", "DNS results can vary based on caching and which DNS resolver is being queried. This tool queries authoritative or public resolvers directly, which may differ from your local ISP's cached results."),
    ("How long does DNS propagation take?", "Typically a few minutes to 48 hours, depending on the TTL (time-to-live) set on the record and how widely it has been cached across resolvers."),
    ("What's the difference between A and CNAME records?", "An A record points a domain directly to an IPv4 address. A CNAME record points a domain to another domain name instead of an IP."),
    ("Why is my MX record important?", "MX records tell other mail servers where to deliver email for your domain. Missing or incorrect MX records cause email delivery failures."),
],
"down-checker.html": [
    ("What does 'down for everyone' mean?", "It means the site failed to respond from our server's vantage point, which is separate from your own network — indicating the problem is likely on the server or hosting side, not your connection."),
    ("Why might a site be down for me but not others?", "This can happen due to local DNS caching issues, ISP-level routing problems, or firewall/network restrictions specific to your connection."),
    ("What's a common cause of a site going down?", "Server crashes, exceeded resource limits, expired SSL certificates, DNS misconfiguration, or hosting provider outages are all common causes."),
    ("Does this tool tell me why a site is down?", "It confirms reachability, not the root cause. Pair it with an HTTP header check, SSL checker, or DNS lookup to narrow down the specific issue."),
],
"http-header-checker.html": [
    ("What are the most important security headers?", "Strict-Transport-Security (HSTS), Content-Security-Policy (CSP), X-Frame-Options, and X-Content-Type-Options are widely considered baseline security headers."),
    ("Why don't I see a Server header?", "Many servers are configured to hide or obscure this header intentionally, as a minor security-through-obscurity measure."),
    ("What does Cache-Control tell me?", "It defines how long browsers and proxies should cache a response, which directly affects how quickly your users see updated content."),
    ("Can headers reveal my tech stack?", "Yes — headers like X-Powered-By or Server often reveal the underlying framework or server software, which is why many production sites strip them."),
],
"ip-geolocation.html": [
    ("How accurate is IP geolocation?", "Country-level accuracy is generally high. City-level accuracy varies and can be off by tens of kilometers, especially for mobile or ISP-shared IP ranges."),
    ("Can IP geolocation identify a specific person?", "No. It identifies the approximate location of the ISP or network infrastructure serving that IP, not an individual or exact address."),
    ("Why does the location seem wrong?", "This is common with mobile carriers, VPNs, and corporate networks, which often route traffic through infrastructure located far from the actual user."),
    ("What is ASN in these results?", "The Autonomous System Number identifies the network/organization that owns the IP range, useful for identifying the ISP or hosting provider."),
],
"port-checker.html": [
    ("What does 'filtered' mean versus 'closed'?", "Closed means the port actively refused the connection. Filtered means no response was received at all, usually because a firewall is silently dropping the traffic."),
    ("Is it safe to leave any ports open?", "Only ports required for services you intend to expose publicly (e.g. 80/443 for a web server) should be open. Everything else should be firewalled by default."),
    ("Why is my port closed even though the service is running?", "This usually means a firewall (on the server itself, a cloud provider's security group, or your router) is blocking external access to that port."),
    ("What ports are commonly targeted by attackers?", "22 (SSH), 3389 (RDP), 3306/5432 (databases), and 23 (Telnet) are frequently scanned and targeted if left exposed without proper authentication."),
],
"reverse-dns.html": [
    ("What is a PTR record?", "A PTR (pointer) record maps an IP address back to a hostname — the reverse of a normal DNS lookup, which maps a hostname to an IP."),
    ("Why does my mail server need a PTR record?", "Many mail providers check that the sending IP's PTR record matches or is related to the sending domain, as a basic anti-spam signal. Missing PTR records commonly cause deliverability issues."),
    ("Can anyone set a PTR record?", "No — PTR records are controlled by whoever owns the IP address block, typically your hosting provider or ISP, not the domain owner directly."),
    ("Why does the PTR hostname look unfamiliar?", "Cloud and VPS providers often assign generic PTR hostnames by default (e.g. based on the datacenter). You usually need to request a custom PTR record from your provider."),
],
"speed-test.html": [
    ("Why do my results vary between tests?", "Network conditions fluctuate due to congestion, Wi-Fi interference, and the number of devices sharing your connection at the time of the test."),
    ("What's a good speed for video calls?", "Generally 3-4 Mbps up/down is sufficient for HD video calls, though higher is better for group calls or simultaneous usage."),
    ("Why is my upload speed lower than download?", "Most residential ISP plans are asymmetric by design, prioritizing download bandwidth since most consumer usage is download-heavy (streaming, browsing)."),
    ("What does jitter affect?", "High jitter (inconsistent latency) causes choppy audio/video in calls and lag spikes in online gaming, even if average speed looks fine."),
],
"ssl-checker.html": [
    ("What happens when an SSL certificate expires?", "Browsers will show a security warning and typically block users from proceeding, effectively taking your site offline for most visitors until it's renewed."),
    ("What is a certificate chain issue?", "It means the server isn't sending the full chain of intermediate certificates needed for browsers to verify trust up to a recognized root authority — a very common misconfiguration."),
    ("Is a self-signed certificate a problem?", "For public-facing production sites, yes — browsers won't trust it and will show warnings. Self-signed certs are fine for internal/development use only."),
    ("How often should I renew my SSL certificate?", "Most modern certificates (like Let's Encrypt) are valid for 90 days and are usually configured to auto-renew. Manually issued certificates can be valid for up to a year or more."),
],
"subnet-calculator.html": [
    ("What's the difference between network and broadcast address?", "The network address identifies the subnet itself and can't be assigned to a device. The broadcast address is used to send data to every device on that subnet simultaneously — also not assignable."),
    ("How many usable hosts does a /30 provide?", "A /30 provides 4 total addresses, 2 of which are usable for hosts (network and broadcast take the other 2) — commonly used for point-to-point links."),
    ("What is CIDR notation?", "CIDR (e.g. /24) is a shorthand for expressing how many bits of an IP address are used for the network portion, which determines subnet size."),
    ("Why do cloud providers reserve extra IPs per subnet?", "Providers like AWS reserve a handful of addresses per subnet (beyond network/broadcast) for internal routing and DNS purposes, slightly reducing usable host count."),
],
"what-is-my-ip.html": [
    ("Is my public IP the same as my device's IP?", "No. Your device has a private/local IP (e.g. 192.168.x.x) used within your home network, while your public IP is what the internet sees, usually assigned by your ISP."),
    ("Does my public IP change?", "For most residential connections, yes — ISPs typically assign dynamic IPs that can change periodically or on router restart, unless you've paid for a static IP."),
    ("Is it safe to share my IP address?", "Your IP alone reveals only your approximate location and ISP — not personal identity. Still, avoid sharing it unnecessarily, especially alongside other identifying information."),
    ("Why does my IP show as IPv6 sometimes and IPv4 other times?", "Many networks support both protocols simultaneously (dual-stack), and which one is displayed depends on which protocol your device and our server negotiate for that particular connection."),
],
"whois-lookup.html": [
    ("Why don't I see the owner's name or contact info?", "Most registrars now offer WHOIS privacy protection by default, which masks personal registrant details behind the registrar's own proxy contact information."),
    ("What happens if a domain expires?", "It typically enters a grace period, then a redemption period with extra fees to reclaim it, and finally becomes publicly available for anyone to register."),
    ("What do nameservers in WHOIS tell me?", "They show which DNS provider is managing the domain's records, which can indicate the hosting setup even before doing a full DNS lookup."),
    ("Can WHOIS data be inaccurate?", "Yes — registrants aren't always required to keep information current, and some registrars cache WHOIS data, so it may lag behind recent changes."),
],
}

def build_schema(faqs):
    schema = {
        "@context": "https://schema.org",
        "@type": "FAQPage",
        "mainEntity": [
            {
                "@type": "Question",
                "name": q,
                "acceptedAnswer": {"@type": "Answer", "text": a}
            } for q, a in faqs
        ]
    }
    return json.dumps(schema, indent=2)

results = []
pattern = re.compile(r'<script type="application/ld\+json">.*?</script>', re.DOTALL)

for fname, faqs in FAQS.items():
    path = os.path.join(BASE, fname)
    if not os.path.exists(path):
        results.append((fname, "SKIPPED - not found"))
        continue
    with open(path, "r") as f:
        content = f.read()
    if not pattern.search(content):
        results.append((fname, "SKIPPED - no JSON-LD block found"))
        continue

    new_block = '<script type="application/ld+json">\n' + build_schema(faqs) + '\n</script>'
    content = pattern.sub(lambda m: new_block, content, count=1)

    with open(path, "w") as f:
        f.write(content)
    results.append((fname, "FIXED"))

print("--- Results ---")
for fname, status in results:
    print(f"{fname}: {status}")
