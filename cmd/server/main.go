package main

import (
	"fmt"
	"log"
	"os"

	"traceqube/internal/handlers"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load(".env.api")
	port := os.Getenv("PORT")
	if port == "" {
		port = "5500"
	}

	app := fiber.New(fiber.Config{AppName: "TraceQube v2", BodyLimit: 8 * 1024 * 1024})
	app.Use(logger.New())
	app.Use(recover.New())
	app.Static("/static", "./web/static")
	app.Static("/", "./web/root")

	// Main pages
	app.Get("/", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "home.html", handlers.PageData{Title: "TraceQube - Free Web & Network Tools", Description: "Free online tools to check IPs, DNS records, open ports, HTTP headers, SSL certificates and more.", Canonical: "/"})
	})
	app.Get("/tools", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "tools.html", handlers.PageData{Title: "All Free Network Tools - TraceQube", Description: "Browse all 11 free network and web diagnostic tools by TraceQube.", Canonical: "/tools"})
	})
	app.Get("/about", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "about.html", handlers.PageData{Title: "About TraceQube", Description: "Learn about TraceQube and our free network diagnostic tools.", Canonical: "/about"})
	})
	app.Get("/contact", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "contact.html", handlers.PageData{Title: "Contact Us - TraceQube", Description: "Get in touch with TraceQube.", Canonical: "/contact"})
	})
	app.Get("/privacy-policy", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "privacy-policy.html", handlers.PageData{Title: "Privacy Policy - TraceQube", Description: "Read TraceQube privacy policy.", Canonical: "/privacy-policy"})
	})
	app.Get("/terms", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "terms.html", handlers.PageData{Title: "Terms of Service - TraceQube", Description: "Read TraceQube terms of service.", Canonical: "/terms"})
	})
	app.Get("/pricing", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "pricing.html", handlers.PageData{Title: "Pricing - TraceQube API Plans", Description: "Simple transparent pricing for the TraceQube API.", Canonical: "/pricing"})
	})
	app.Get("/blog", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog.html", handlers.PageData{Title: "Blog - TraceQube", Description: "Network diagnostics and developer guides.", Canonical: "/blog"})
	})
	app.Get("/blog/page/:num", func(c *fiber.Ctx) error {
		num := c.Params("num")
		if num == "1" {
			return c.Redirect("/blog", 301)
		}
		tmplName := fmt.Sprintf("blog-page-%s.html", num)
		if _, err := os.Stat("web/templates/" + tmplName); err != nil {
			return c.Redirect("/blog", 302)
		}
		return handlers.RenderPage(c, tmplName, handlers.PageData{Title: "Blog - Page " + num + " | TraceQube", Description: "Network diagnostics and developer guides.", Canonical: "/blog/page/" + num})
	})
	app.Get("/api-docs", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "api-docs.html", handlers.PageData{Title: "API Docs - TraceQube", Description: "API documentation.", Canonical: "/api-docs"})
	})
	app.Get("/sitemap.xml", func(c *fiber.Ctx) error {
		c.Set("Content-Type", "application/xml")
		return c.SendFile("./web/static/sitemap.xml")
	})
	app.Get("/robots.txt", func(c *fiber.Ctx) error {
		return c.SendFile("./web/static/robots.txt")
	})

	// Blog posts

	// Tool pages
	app.Get("/tools/what-is-my-ip", handlers.WhatIsMyIPPage)
	app.Get("/tools/dns-lookup", handlers.DNSLookupPage)
	app.Get("/tools/ping-test", handlers.PingTestPage)
	app.Get("/tools/triage", handlers.TriagePage)
	app.Get("/tools/port-checker", handlers.PortCheckerPage)
	app.Get("/tools/http-header-checker", handlers.HTTPHeaderPage)
	app.Get("/tools/whois-lookup", handlers.WhoisPage)
	app.Get("/tools/ssl-checker", handlers.SSLCheckerPage)
	app.Get("/tools/ip-geolocation", handlers.IPGeoPage)
	app.Get("/tools/reverse-dns", handlers.ReverseDNSPage)
	app.Get("/tools/subnet-calculator", handlers.SubnetPage)
	app.Get("/tools/down-checker", handlers.DownCheckerPage)
	app.Get("/tools/speed-test", handlers.SpeedTestPage)





























































































































	app.Get("/blog/dns-spoofing-prevention", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-dns-spoofing-prevention.html", handlers.PageData{Title: "Understanding DNS Spoofing and its Prevention | TraceQube", Description: "Learn how DNS spoofing works and the various methods to prevent it from compromising your network security.", Canonical: "/blog/dns-spoofing-prevention"})
	})

	app.Get("/blog/ipv6-importance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ipv6-importance.html", handlers.PageData{Title: "The Importance of IPv6 in Modern Networking | TraceQube", Description: "Discover the significance of IPv6 in today's network infrastructure and its impact on future-proofing your network.", Canonical: "/blog/ipv6-importance"})
	})

	app.Get("/blog/ssl-tls-configuration", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ssl-tls-configuration.html", handlers.PageData{Title: "SSL/TLS Configuration for Secure Web Applications | TraceQube", Description: "Find out the best practices for configuring SSL/TLS certificates to ensure secure communication between web applications and clients.", Canonical: "/blog/ssl-tls-configuration"})
	})

	app.Get("/blog/tcp-tuning-for-performance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-tcp-tuning-for-performance.html", handlers.PageData{Title: "Optimizing Server Response Time with TCP Tuning | TraceQube", Description: "Explore the benefits of fine-tuning TCP settings to improve server response times and overall web application performance.", Canonical: "/blog/tcp-tuning-for-performance"})
	})

	app.Get("/blog/internet-protocol-suites", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-internet-protocol-suites.html", handlers.PageData{Title: "A Comprehensive Guide to Internet Protocol Suites | TraceQube", Description: "Get a detailed overview of various internet protocol suites and their applications in modern network architectures.", Canonical: "/blog/internet-protocol-suites"})
	})

	app.Get("/blog/optimizing-network-performance-with-qos-policies", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-optimizing-network-performance-with-qos-policies.html", handlers.PageData{Title: "Optimizing Network Performance with QoS Policies | TraceQube", Description: "Learn how to implement Quality of Service policies to prioritize critical network traffic and boost overall performance.", Canonical: "/blog/optimizing-network-performance-with-qos-policies"})
	})

	app.Get("/blog/understanding-dnssec-a-guide-for-network-administrators", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-understanding-dnssec-a-guide-for-network-administrators.html", handlers.PageData{Title: "Understanding DNSSEC: A Guide for Network Administrators | TraceQube", Description: "Discover the benefits and best practices of implementing DNSSEC to secure your DNS infrastructure.", Canonical: "/blog/understanding-dnssec-a-guide-for-network-administrators"})
	})

	app.Get("/blog/mastering-ip-address-management-strategies-for-efficient-ip-address-allocation", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-mastering-ip-address-management-strategies-for-efficient-ip-address-allocation.html", handlers.PageData{Title: "Mastering IP Address Management: Strategies for Efficient IP Address Allocation | TraceQube", Description: "Get expert advice on managing IP addresses effectively to reduce waste and ensure optimal network utilization.", Canonical: "/blog/mastering-ip-address-management-strategies-for-efficient-ip-address-allocation"})
	})

	app.Get("/blog/ssl-tls-certificate-chain-validation-best-practices-and-common-pitfalls", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ssl-tls-certificate-chain-validation-best-practices-and-common-pitfalls.html", handlers.PageData{Title: "SSL/TLS Certificate Chain Validation: Best Practices and Common Pitfalls | TraceQube", Description: "Understand the importance of proper SSL/TLS certificate chain validation and how to avoid common mistakes.", Canonical: "/blog/ssl-tls-certificate-chain-validation-best-practices-and-common-pitfalls"})
	})

	app.Get("/blog/web-server-optimization-techniques-for-improving-http-2-performance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-web-server-optimization-techniques-for-improving-http-2-performance.html", handlers.PageData{Title: "Web Server Optimization: Techniques for Improving HTTP/2 Performance | TraceQube", Description: "Discover the latest techniques for optimizing HTTP/2 performance and improving web server efficiency.", Canonical: "/blog/web-server-optimization-techniques-for-improving-http-2-performance"})
	})

	app.Get("/blog/dns-propagation-times", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-dns-propagation-times.html", handlers.PageData{Title: "Understanding DNS Propagation Times | TraceQube", Description: "Learn how DNS propagation times impact your website's availability and what you can do to minimize the effects.", Canonical: "/blog/dns-propagation-times"})
	})

	app.Get("/blog/ssl-tls-certificate-chain-of-trust", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ssl-tls-certificate-chain-of-trust.html", handlers.PageData{Title: "SSL/TLS Certificate Chain of Trust | TraceQube", Description: "Discover the process behind how SSL/TLS certificates verify the identity of websites and protect user data.", Canonical: "/blog/ssl-tls-certificate-chain-of-trust"})
	})

	app.Get("/blog/ip-address-allocation-guide", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ip-address-allocation-guide.html", handlers.PageData{Title: "IP Address Allocation: A Guide | TraceQube", Description: "Get a comprehensive overview of the IP address allocation process and how it affects network communication.", Canonical: "/blog/ip-address-allocation-guide"})
	})

	app.Get("/blog/web-performance-optimization-caching", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-web-performance-optimization-caching.html", handlers.PageData{Title: "Web Performance Optimization: The Importance of Caching | TraceQube", Description: "Learn how caching can significantly improve your website's loading speed and user experience.", Canonical: "/blog/web-performance-optimization-caching"})
	})

	app.Get("/blog/server-administration-monitoring-log-files", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-server-administration-monitoring-log-files.html", handlers.PageData{Title: "Server Administration: Best Practices for Monitoring Log Files | TraceQube", Description: "Discover the best practices for monitoring log files to identify potential issues and improve server security.", Canonical: "/blog/server-administration-monitoring-log-files"})
	})

	app.Get("/blog/optimizing-dns-for-website-performance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-optimizing-dns-for-website-performance.html", handlers.PageData{Title: "Optimizing DNS for Improved Website Performance | TraceQube", Description: "Implementing efficient DNS techniques can significantly boost your website's load times and user experience.", Canonical: "/blog/optimizing-dns-for-website-performance"})
	})

	app.Get("/blog/the-importance-of-ip-address-management-in-cloud-environments", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-the-importance-of-ip-address-management-in-cloud-environments.html", handlers.PageData{Title: "The Importance of IP Address Management in Cloud Environments | TraceQube", Description: "Proper IP address management is crucial for maintaining network security and scalability in cloud-based systems.", Canonical: "/blog/the-importance-of-ip-address-management-in-cloud-environments"})
	})

	app.Get("/blog/the-role-of-ssl-tls-in-secure-web-browsing", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-the-role-of-ssl-tls-in-secure-web-browsing.html", handlers.PageData{Title: "Understanding the Role of SSL/TLS in Secure Web Browsing | TraceQube", Description: "SSL/TLS certificates are the backbone of secure web communication, ensuring that sensitive data remains encrypted and protected from cyber threats.", Canonical: "/blog/the-role-of-ssl-tls-in-secure-web-browsing"})
	})

	app.Get("/blog/troubleshooting-common-network-connectivity-issues-with-traceqube", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-troubleshooting-common-network-connectivity-issues-with-traceqube.html", handlers.PageData{Title: "Troubleshooting Common Network Connectivity Issues with TraceQube | TraceQube", Description: "Our network diagnostics tool helps you identify and resolve network connectivity problems, minimizing downtime and improving overall system reliability.", Canonical: "/blog/troubleshooting-common-network-connectivity-issues-with-traceqube"})
	})

	app.Get("/blog/a-beginners-guide-to-internet-protocols-and-tcp-ip", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-a-beginners-guide-to-internet-protocols-and-tcp-ip.html", handlers.PageData{Title: "A Beginner's Guide to Understanding Internet Protocols and TCP/IP | TraceQube", Description: "In this article, we break down the basics of internet protocols and TCP/IP, providing a comprehensive overview for developers and network administrators alike.", Canonical: "/blog/a-beginners-guide-to-internet-protocols-and-tcp-ip"})
	})

	app.Get("/blog/dns-threats-and-vulnerabilities", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-dns-threats-and-vulnerabilities.html", handlers.PageData{Title: "Exploring the Dark Side of DNS: Threats and Vulnerabilities | TraceQube", Description: "DNS is a critical component of the internet, but it's also a common target for cyber attacks and exploitation.", Canonical: "/blog/dns-threats-and-vulnerabilities"})
	})

	app.Get("/blog/ip-address-classes-for-beginners", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ip-address-classes-for-beginners.html", handlers.PageData{Title: "A Beginner's Guide to Understanding IP Address Classes | TraceQube", Description: "This article explains the different types of IP address classes and how they are used in networking.", Canonical: "/blog/ip-address-classes-for-beginners"})
	})

	app.Get("/blog/ssl-tls-optimization-for-web-performance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ssl-tls-optimization-for-web-performance.html", handlers.PageData{Title: "Optimizing SSL/TLS Configuration for Improved Web Performance | TraceQube", Description: "Improperly configured SSL/TLS can lead to slow load times and decreased user engagement.", Canonical: "/blog/ssl-tls-optimization-for-web-performance"})
	})

	app.Get("/blog/server-maintenance-for-web-security", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-server-maintenance-for-web-security.html", handlers.PageData{Title: "The Importance of Regular Server Maintenance for Web Security | TraceQube", Description: "Regular server maintenance is crucial for preventing security vulnerabilities and ensuring web applications remain secure.", Canonical: "/blog/server-maintenance-for-web-security"})
	})

	app.Get("/blog/http-3-next-generation-protocols", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-http-3-next-generation-protocols.html", handlers.PageData{Title: "Understanding HTTP/3: The Next Generation of Internet Protocols | TraceQube", Description: "HTTP/3 is a new internet protocol designed to improve performance and reliability in web applications.", Canonical: "/blog/http-3-next-generation-protocols"})
	})

	app.Get("/blog/dns-cache-expiration-impact-on-website-performance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-dns-cache-expiration-impact-on-website-performance.html", handlers.PageData{Title: "Understanding the Impact of DNS Cache Expiration on Website Performance | TraceQube", Description: "Learn how DNS cache expiration affects website performance and discover strategies to optimize your DNS settings.", Canonical: "/blog/dns-cache-expiration-impact-on-website-performance"})
	})

	app.Get("/blog/ip-address-scanning-for-network-security", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ip-address-scanning-for-network-security.html", handlers.PageData{Title: "The Importance of Regular IP Address Scanning for Network Security | TraceQube", Description: "Discover why regular IP address scanning is essential for identifying vulnerabilities and strengthening your network security.", Canonical: "/blog/ip-address-scanning-for-network-security"})
	})

	app.Get("/blog/ssl-tls-certificates-for-secure-web-browsing", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ssl-tls-certificates-for-secure-web-browsing.html", handlers.PageData{Title: "A Beginner's Guide to SSL/TLS Certificates and Secure Web Browsing | TraceQube", Description: "Get started with SSL/TLS certificates and learn how to ensure secure web browsing for your users.", Canonical: "/blog/ssl-tls-certificates-for-secure-web-browsing"})
	})

	app.Get("/blog/troubleshooting-network-connectivity-with-traceroute", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-troubleshooting-network-connectivity-with-traceroute.html", handlers.PageData{Title: "Troubleshooting Common Network Connectivity Issues Using Traceroute | TraceQube", Description: "Learn how to use traceroute to identify and troubleshoot common network connectivity issues.", Canonical: "/blog/troubleshooting-network-connectivity-with-traceroute"})
	})

	app.Get("/blog/ipv6-in-modern-network-architecture", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ipv6-in-modern-network-architecture.html", handlers.PageData{Title: "The Role of Internet Protocol Version 6 (IPv6) in Modern Network Architecture | TraceQube", Description: "Discover the benefits and challenges of implementing IPv6 in your modern network architecture.", Canonical: "/blog/ipv6-in-modern-network-architecture"})
	})

	app.Get("/blog/ipv6-adoption-network-infrastructure", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ipv6-adoption-network-infrastructure.html", handlers.PageData{Title: "The Impact of IPv6 Adoption on Network Infrastructure | TraceQube", Description: "As IPv6 adoption continues to grow, network administrators must be prepared to adapt and update their infrastructure to support the new protocol.", Canonical: "/blog/ipv6-adoption-network-infrastructure"})
	})

	app.Get("/blog/dns-over-https-dot", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-dns-over-https-dot.html", handlers.PageData{Title: "A Deep Dive into DNS over HTTPS (DoH) and DoT | TraceQube", Description: "Discover how DNS over HTTPS and DNS over TLS can improve DNS security and performance in your network.", Canonical: "/blog/dns-over-https-dot"})
	})

	app.Get("/blog/ssl-tls-certificate-validation", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ssl-tls-certificate-validation.html", handlers.PageData{Title: "The Importance of SSL/TLS Certificate Validation in Web Security | TraceQube", Description: "Learn why proper SSL/TLS certificate validation is crucial for ensuring the security and trust of your website and web applications.", Canonical: "/blog/ssl-tls-certificate-validation"})
	})

	app.Get("/blog/http-3-web-performance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-http-3-web-performance.html", handlers.PageData{Title: "Understanding HTTP/3 and its Impact on Web Performance | TraceQube", Description: "Explore the features and benefits of HTTP/3 and how it can improve the speed and reliability of your web applications.", Canonical: "/blog/http-3-web-performance"})
	})

	app.Get("/blog/server-administration-network-load-management", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-server-administration-network-load-management.html", handlers.PageData{Title: "Server Administration Best Practices for Managing Network Load | TraceQube", Description: "Discover essential server administration best practices for effectively managing network load and ensuring the performance of your servers.", Canonical: "/blog/server-administration-network-load-management"})
	})

	app.Get("/blog/optimizing-network-performance-with-path-mtu-discovery", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-optimizing-network-performance-with-path-mtu-discovery.html", handlers.PageData{Title: "Optimizing Network Performance with Path MTU Discovery | TraceQube", Description: "Learn how to optimize network performance by implementing Path MTU discovery to prevent packet fragmentation and reduce latency.", Canonical: "/blog/optimizing-network-performance-with-path-mtu-discovery"})
	})

	app.Get("/blog/the-importance-of-dnssec-for-secure-domain-name-resolution", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-the-importance-of-dnssec-for-secure-domain-name-resolution.html", handlers.PageData{Title: "The Importance of DNSSEC for Secure Domain Name Resolution | TraceQube", Description: "Discover the benefits of using DNSSEC to secure domain name resolution and prevent DNS spoofing attacks.", Canonical: "/blog/the-importance-of-dnssec-for-secure-domain-name-resolution"})
	})

	app.Get("/blog/understanding-ip-address-allocation-and-management", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-understanding-ip-address-allocation-and-management.html", handlers.PageData{Title: "Understanding IP Address Allocation and Management | TraceQube", Description: "Learn about the different types of IP address allocation and management methods used in modern networking environments.", Canonical: "/blog/understanding-ip-address-allocation-and-management"})
	})

	app.Get("/blog/implementing-ssl-tls-certificate-pinning-for-enhanced-web-security", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-implementing-ssl-tls-certificate-pinning-for-enhanced-web-security.html", handlers.PageData{Title: "Implementing SSL/TLS Certificate Pinning for Enhanced Web Security | TraceQube", Description: "Find out how to implement SSL/TLS certificate pinning to prevent man-in-the-middle attacks and enhance web security.", Canonical: "/blog/implementing-ssl-tls-certificate-pinning-for-enhanced-web-security"})
	})

	app.Get("/blog/analyzing-web-performance-metrics-with-http-2-and-tcp", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-analyzing-web-performance-metrics-with-http-2-and-tcp.html", handlers.PageData{Title: "Analyzing Web Performance Metrics with HTTP/2 and TCP | TraceQube", Description: "Learn how to analyze web performance metrics using HTTP/2 and TCP to optimize server and network configurations.", Canonical: "/blog/analyzing-web-performance-metrics-with-http-2-and-tcp"})
	})

	app.Get("/blog/the-impact-of-ipv6-on-network-infrastructure", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-the-impact-of-ipv6-on-network-infrastructure.html", handlers.PageData{Title: "The Impact of IPv6 on Network Infrastructure | TraceQube", Description: "", Canonical: "/blog/the-impact-of-ipv6-on-network-infrastructure"})
	})

	app.Get("/blog/hardening-dns-resolvers-for-enhanced-security", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-hardening-dns-resolvers-for-enhanced-security.html", handlers.PageData{Title: "Hardening DNS Resolvers for Enhanced Security | TraceQube", Description: "", Canonical: "/blog/hardening-dns-resolvers-for-enhanced-security"})
	})

	app.Get("/blog/understanding-tcp-handshakes-for-optimal-web-performance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-understanding-tcp-handshakes-for-optimal-web-performance.html", handlers.PageData{Title: "Understanding TCP Handshakes for Optimal Web Performance | TraceQube", Description: "", Canonical: "/blog/understanding-tcp-handshakes-for-optimal-web-performance"})
	})

	app.Get("/blog/mastering-ssl-tls-configuration-for-secure-web-servers", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-mastering-ssl-tls-configuration-for-secure-web-servers.html", handlers.PageData{Title: "Mastering SSL/TLS Configuration for Secure Web Servers | TraceQube", Description: "", Canonical: "/blog/mastering-ssl-tls-configuration-for-secure-web-servers"})
	})

	app.Get("/blog/a-deep-dive-into-http-2-for-faster-web-applications", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-a-deep-dive-into-http-2-for-faster-web-applications.html", handlers.PageData{Title: "A Deep Dive into HTTP/2 for Faster Web Applications | TraceQube", Description: "", Canonical: "/blog/a-deep-dive-into-http-2-for-faster-web-applications"})
	})

	app.Get("/blog/the-role-of-dns-in-web-application-security", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-the-role-of-dns-in-web-application-security.html", handlers.PageData{Title: "The Role of DNS in Web Application Security | TraceQube", Description: "", Canonical: "/blog/the-role-of-dns-in-web-application-security"})
	})

	app.Get("/blog/optimizing-server-configuration-for-maximum-performance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-optimizing-server-configuration-for-maximum-performance.html", handlers.PageData{Title: "Optimizing Server Configuration for Maximum Performance | TraceQube", Description: "", Canonical: "/blog/optimizing-server-configuration-for-maximum-performance"})
	})

	app.Get("/blog/the-benefits-of-using-load-balancing-for-scalable-applications", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-the-benefits-of-using-load-balancing-for-scalable-applications.html", handlers.PageData{Title: "The Benefits of Using Load Balancing for Scalable Applications | TraceQube", Description: "", Canonical: "/blog/the-benefits-of-using-load-balancing-for-scalable-applications"})
	})

	app.Get("/blog/best-practices-for-secure-server-administration", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-best-practices-for-secure-server-administration.html", handlers.PageData{Title: "Best Practices for Secure Server Administration | TraceQube", Description: "", Canonical: "/blog/best-practices-for-secure-server-administration"})
	})

	app.Get("/blog/understanding-http-status-codes-for-better-web-development", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-understanding-http-status-codes-for-better-web-development.html", handlers.PageData{Title: "Understanding HTTP Status Codes for Better Web Development | TraceQube", Description: "", Canonical: "/blog/understanding-http-status-codes-for-better-web-development"})
	})

	app.Get("/blog/optimizing-dns-resolution", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-optimizing-dns-resolution.html", handlers.PageData{Title: "Optimizing DNS Resolution for Faster Network Performance | TraceQube", Description: "Discover how to improve DNS resolution times and boost network speed with advanced techniques and tools.", Canonical: "/blog/optimizing-dns-resolution"})
	})

	app.Get("/blog/ssl-tls-certificate-hierarchy", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ssl-tls-certificate-hierarchy.html", handlers.PageData{Title: "A Deep Dive into SSL/TLS Certificate Hierarchy and Trust Chains | TraceQube", Description: "Explore the complexities of SSL/TLS certificate hierarchies and trust chains to ensure secure online transactions.", Canonical: "/blog/ssl-tls-certificate-hierarchy"})
	})

	app.Get("/blog/ipv6-modern-networks", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ipv6-modern-networks.html", handlers.PageData{Title: "The Impact of IPv6 on Modern Network Architecture | TraceQube", Description: "Learn how IPv6 is transforming network architecture and what changes you need to make to stay ahead.", Canonical: "/blog/ipv6-modern-networks"})
	})

	app.Get("/blog/server-administration-best-practices", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-server-administration-best-practices.html", handlers.PageData{Title: "Server Administration Best Practices for Enhanced Security and Stability | TraceQube", Description: "Discover the essential server administration best practices to maintain a secure and stable online environment.", Canonical: "/blog/server-administration-best-practices"})
	})

	app.Get("/blog/web-performance-metrics", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-web-performance-metrics.html", handlers.PageData{Title: "Understanding Web Performance Metrics: Page Load Time, Bounce Rate, and More | TraceQube", Description: "Master the key web performance metrics to optimize your website's load time, user experience, and online success.", Canonical: "/blog/web-performance-metrics"})
	})

	app.Get("/blog/dns-amplification-attacks-prevention", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-dns-amplification-attacks-prevention.html", handlers.PageData{Title: "Protecting Against DNS Amplification Attacks | TraceQube", Description: "Learn how to prevent DNS amplification attacks and protect your network from these types of cyber threats.", Canonical: "/blog/dns-amplification-attacks-prevention"})
	})

	app.Get("/blog/http1-1-vs-http2", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-http1-1-vs-http2.html", handlers.PageData{Title: "Understanding the Difference Between HTTP/1.1 and HTTP/2 | TraceQube", Description: "Discover the key differences between HTTP/1.1 and HTTP/2, and how they impact web performance.", Canonical: "/blog/http1-1-vs-http2"})
	})

	app.Get("/blog/server-configuration-web-performance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-server-configuration-web-performance.html", handlers.PageData{Title: "Optimizing Server Configuration for Improved Web Performance | TraceQube", Description: "Learn how to optimize server configuration to improve web performance and reduce latency.", Canonical: "/blog/server-configuration-web-performance"})
	})

	app.Get("/blog/ssl-tls-certificate-management", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ssl-tls-certificate-management.html", handlers.PageData{Title: "SSL/TLS Certificate Management Best Practices | TraceQube", Description: "Get the best practices for managing SSL/TLS certificates to ensure secure web communications.", Canonical: "/blog/ssl-tls-certificate-management"})
	})

	app.Get("/blog/understanding-dns-query-types", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-understanding-dns-query-types.html", handlers.PageData{Title: "Understanding DNS Query Types: A Deep Dive into A, AAAA, and NS Records | TraceQube", Description: "Learn about the different types of DNS query records and how they impact your network's performance.", Canonical: "/blog/understanding-dns-query-types"})
	})

	app.Get("/blog/common-ip-addressing-misconfigurations", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-common-ip-addressing-misconfigurations.html", handlers.PageData{Title: "Uncovering Common Misconfigurations in IP Addressing | TraceQube", Description: "Identify and rectify common mistakes in IP address configuration to ensure network stability and security.", Canonical: "/blog/common-ip-addressing-misconfigurations"})
	})

	app.Get("/blog/understanding-dns-resolution", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-understanding-dns-resolution.html", handlers.PageData{Title: "Understanding the Basics of DNS Resolution | TraceQube", Description: "DNS resolution is the process by which a domain name is converted into an IP address.", Canonical: "/blog/understanding-dns-resolution"})
	})

	app.Get("/blog/ssl-tls-certificates-for-web-security", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ssl-tls-certificates-for-web-security.html", handlers.PageData{Title: "The Importance of SSL/TLS Certificates for Web Security | TraceQube", Description: "SSL/TLS certificates are crucial for establishing trust and security between a website and its users.", Canonical: "/blog/ssl-tls-certificates-for-web-security"})
	})

	app.Get("/blog/troubleshooting-ip-addressing-issues", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-troubleshooting-ip-addressing-issues.html", handlers.PageData{Title: "Troubleshooting Common IP Addressing Issues | TraceQube", Description: "IP addressing issues can lead to connectivity problems and network downtime if not addressed promptly.", Canonical: "/blog/troubleshooting-ip-addressing-issues"})
	})

	app.Get("/blog/http2-protocol-benefits", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-http2-protocol-benefits.html", handlers.PageData{Title: "A Deep Dive into the HTTP/2 Protocol and Its Benefits | TraceQube", Description: "The HTTP/2 protocol offers several benefits over its predecessor, including improved performance and reduced latency.", Canonical: "/blog/http2-protocol-benefits"})
	})

	app.Get("/blog/ipv6-impact-on-network-architecture", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ipv6-impact-on-network-architecture.html", handlers.PageData{Title: "The Impact of IPv6 on Modern Network Architecture | TraceQube", Description: "IPv6 is slowly replacing IPv4, but what does this mean for network administrators and engineers?", Canonical: "/blog/ipv6-impact-on-network-architecture"})
	})

	app.Get("/blog/dns-vs-dhcp", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-dns-vs-dhcp.html", handlers.PageData{Title: "Understanding the Difference Between DNS and DHCP | TraceQube", Description: "Discover how DNS and DHCP work together to provide internet access to devices.", Canonical: "/blog/dns-vs-dhcp"})
	})

	app.Get("/blog/load-balancing-for-high-traffic-web-applications", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-load-balancing-for-high-traffic-web-applications.html", handlers.PageData{Title: "The Importance of Load Balancing in High-Traffic Web Applications | TraceQube", Description: "Discover how load balancing can improve the scalability and reliability of web applications.", Canonical: "/blog/load-balancing-for-high-traffic-web-applications"})
	})

	app.Get("/blog/network-latency-and-web-performance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-network-latency-and-web-performance.html", handlers.PageData{Title: "Analyzing Network Latency and Its Impact on Web Performance | TraceQube", Description: "Learn how to identify and mitigate network latency to improve web application responsiveness.", Canonical: "/blog/network-latency-and-web-performance"})
	})

	app.Get("/blog/tcp-ip-protocol-stacks-and-network-communication", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-tcp-ip-protocol-stacks-and-network-communication.html", handlers.PageData{Title: "A Deep Dive into TCP/IP Protocol Stacks and Their Role in Network Communication | TraceQube", Description: "Explore the inner workings of TCP/IP protocol stacks and how they enable network communication.", Canonical: "/blog/tcp-ip-protocol-stacks-and-network-communication"})
	})

	app.Get("/blog/securing-web-servers-against-common-vulnerabilities", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-securing-web-servers-against-common-vulnerabilities.html", handlers.PageData{Title: "Best Practices for Securing Web Servers Against Common Vulnerabilities | TraceQube", Description: "Discover essential security measures to protect web servers from common attacks and vulnerabilities.", Canonical: "/blog/securing-web-servers-against-common-vulnerabilities"})
	})

	app.Get("/blog/dnssec-guide-for-domain-name-system-security", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-dnssec-guide-for-domain-name-system-security.html", handlers.PageData{Title: "Mastering the Art of DNSSEC: A Guide to Domain Name System Security | TraceQube", Description: "Learn the fundamentals of DNSSEC and how it can enhance the security of your domain name system.", Canonical: "/blog/dnssec-guide-for-domain-name-system-security"})
	})

	app.Get("/blog/optimizing-http2-for-web-application-performance", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-optimizing-http2-for-web-application-performance.html", handlers.PageData{Title: "Optimizing HTTP/2 for Improved Web Application Performance | TraceQube", Description: "Discover the benefits of HTTP/2 and how to optimize it for improved web application performance.", Canonical: "/blog/optimizing-http2-for-web-application-performance"})
	})

	app.Get("/blog/ntp-and-time-synchronization-across-networks", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-ntp-and-time-synchronization-across-networks.html", handlers.PageData{Title: "The Role of NTP in Maintaining Accurate Time Synchronization Across Networks | TraceQube", Description: "Learn about the importance of NTP in maintaining accurate time synchronization across networks.", Canonical: "/blog/ntp-and-time-synchronization-across-networks"})
	})

	app.Get("/blog/understanding-bgp-communities-and-route-optimization", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-understanding-bgp-communities-and-route-optimization.html", handlers.PageData{Title: "Understanding BGP Communities and Route Optimization | TraceQube", Description: "BGP communities are a powerful tool for route optimization, but require a deep understanding of BGP routing.", Canonical: "/blog/understanding-bgp-communities-and-route-optimization"})
	})

	app.Get("/blog/the-importance-of-dnssec-for-domain-security", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-the-importance-of-dnssec-for-domain-security.html", handlers.PageData{Title: "The Importance of DNSSEC for Domain Security | TraceQube", Description: "DNSSEC is a critical component of domain security, providing authentication and integrity for DNS records.", Canonical: "/blog/the-importance-of-dnssec-for-domain-security"})
	})

	app.Get("/blog/protecting-against-ssl-tls-vulnerabilities-with-certificate-pinning", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-protecting-against-ssl-tls-vulnerabilities-with-certificate-pinning.html", handlers.PageData{Title: "Protecting Against SSL/TLS Vulnerabilities with Certificate Pinning | TraceQube", Description: "Certificate pinning is a powerful security technique for protecting against SSL/TLS vulnerabilities and man-in-the-middle attacks.", Canonical: "/blog/protecting-against-ssl-tls-vulnerabilities-with-certificate-pinning"})
	})

	app.Get("/blog/monitoring-and-troubleshooting-with-icmpv6-and-neighbor-discovery", func(c *fiber.Ctx) error {
		return handlers.RenderPage(c, "blog-monitoring-and-troubleshooting-with-icmpv6-and-neighbor-discovery.html", handlers.PageData{Title: "Monitoring and Troubleshooting with ICMPv6 and Neighbor Discovery | TraceQube", Description: "ICMPv6 and neighbor discovery are essential tools for monitoring and troubleshooting IPv6 network connectivity issues.", Canonical: "/blog/monitoring-and-troubleshooting-with-icmpv6-and-neighbor-discovery"})
	})
	// API endpoints
	app.Get("/api/ip", handlers.GetMyIP)
	app.Get("/api/dns", handlers.DNSLookup)
	app.Get("/api/ping", handlers.PingTest)
	app.Get("/api/port", handlers.PortCheck)
	app.Get("/api/http-headers", handlers.HTTPHeaderCheck)
	app.Get("/api/whois", handlers.WhoisLookup)
	app.Get("/api/ssl", handlers.SSLCheck)
	app.Get("/api/geo", handlers.IPGeoLookup)
	app.Get("/api/reverse-dns", handlers.ReverseDNS)
	app.Get("/api/subnet", handlers.SubnetCalc)
	app.Get("/api/down", handlers.DownCheck)
	app.Post("/api/ai/analyze", handlers.AIAnalyze)
	app.Post("/api/ai/triage", handlers.AITriage)


	// 404 handler
	app.Use(func(c *fiber.Ctx) error {
		c.Status(404)
		return handlers.RenderPage(c, "404.html", handlers.PageData{
			Title:       "404 - Page Not Found | TraceQube",
			Description: "The page you are looking for does not exist.",
			Canonical:   "/",
		})
	})
	log.Printf("TraceQube starting on port %s", port)
	log.Fatal(app.Listen(":" + port))
}
