---
title: Records
layout: layout.vto
---
# exists.lol

## Record types

| Record type | What it does | Use it for | Example value |
|---|---|---|---|
| `A` | Points your subdomain to an IPv4 address. | VPS, dedicated server, self-hosted website, API server. | `1.2.3.4` |
| `AAAA` | Points your subdomain to an IPv6 address. | VPS or server with IPv6 support. | `2606:4700:4700::1111` |
| `CNAME` | Points your subdomain to another hostname. | GitHub Pages, Vercel, Netlify, Railway, Render, custom hosting. | `username.github.io` |
| `TXT` | Stores text data in DNS. | Verification records, ownership checks, custom metadata. | `hello-from-exists-lol` |
| `MX` | Points mail delivery to a mail server. | Email hosting for your subdomain. | `10 mail.example.com` |

> Note: `CNAME` cannot be mixed with `A`, `AAAA`, `TXT`, or `MX` for the same subdomain.
