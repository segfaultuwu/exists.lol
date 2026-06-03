---
title: Rules - exists.lol
layout: layout.vto
---

# Rules

These rules apply to every `*.exists.lol` subdomain request.

## General rules

- Do not use `exists.lol` for phishing, malware, scams, spam, botnets, or illegal content.
- Do not impersonate companies, projects, services, public figures, or other people.
- Do not host content that is designed to steal accounts, tokens, passwords, cookies, sessions, or private data.
- Do not use misleading names that make your subdomain look official if it is not.
- Do not abuse the service, automation, GitHub Actions, DNS, or the review process.
- Keep your subdomain useful, safe, and reasonable.

## Subdomain rules

Your subdomain must:

- use only lowercase letters, numbers, and dashes;
- not start or end with a dash;
- not contain spaces, underscores, dots, or special characters;
- not be longer than 63 characters;
- not be a reserved name.

Valid examples:

- `segfault.exists.lol`
- `my-project.exists.lol`
- `bot123.exists.lol`

Invalid examples:

- `MyProject.exists.lol`
- `my_project.exists.lol`
- `-project.exists.lol`
- `project-.exists.lol`
- `my.project.exists.lol`

## Reserved names

The following names are reserved and cannot be registered:

- `www`
- `mail`
- `api`
- `admin`
- `root`
- `support`
- `ns1`
- `ns2`
- `ftp`
- `dashboard`
- `status`
- `cdn`
- `assets`
- `login`
- `auth`
- `account`
- `accounts`
- `billing`
- `security`

More names may be reserved later if needed.

## DNS record rules

Supported record types:

- `A`
- `AAAA`
- `CNAME`
- `TXT`
- `MX`

Rules:

- Do not use wildcard records.
- Do not mix `CNAME` with other record types for the same subdomain.
- Do not submit empty record values.
- Do not submit fake, broken, or intentionally misleading DNS values.
- Your DNS target must be controlled by you or used with permission.

## Content rules

Your subdomain may be removed if it points to:

- phishing pages;
- malware downloads;
- scam pages;
- spam or abuse infrastructure;
- token grabbers;
- fake login pages;
- impersonation pages;
- illegal content;
- content that causes harm to users or services.

## Pull request rules

When opening a pull request:

- Add or edit only files related to your subdomain request.
- Put your domain file inside the `domains` folder.
- Make sure the file name matches the requested subdomain.
- Make sure your JSON is valid.
- Fill out the pull request template.
- Do not modify other users' domain files unless you have permission.
- Do not modify project code, workflows, docs, or config in a domain request PR.

## Removal policy

A subdomain may be removed without notice if it breaks these rules, becomes abusive, points to dangerous content, or causes operational issues for the project.

If your subdomain was removed by mistake, open an issue or contact the maintainers on Discord.

> Note:
> exists.lol is free and community-maintained. Use it responsibly so it can stay available for everyone.
