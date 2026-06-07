import type { DomainRecord, RecordType } from "./schema";

const RESERVED = new Set([
  "www",
  "mail",
  "api",
  "admin",
  "root",
  "support",
  "ns1",
  "ns2",
  "ftp",
  "dashboard",
  "status",
  "cdn",
  "assets",
  "login",
  "auth",
  "account",
  "accounts",
  "billing",
]);

export function validateSubdomain(name: string): string | null {
  if (!name) return "subdomain is required";

  if (name.length > 253) return "domain name is too long";

  const labels = name.split(".");

  for (const label of labels) {
    const error = validateDnsLabel(label);
    if (error) return error;
  }

  const root = labels.at(-1);

  if (!root) {
    return "empty domain name";
  }

  if (RESERVED.has(root)) {
    return "subdomain is reserved";
  }

  return null;
}

function validateDnsLabel(label: string): string | null {
  if (!label) return "empty DNS label";

  if (label.length > 63) return "DNS label is too long";

  if (/^[a-z0-9-]+$/.test(label)) {
    if (label.startsWith("-") || label.endsWith("-")) {
      return "DNS label cannot start or end with dash";
    }

    return null;
  }

  if (/^_[a-z0-9-]+$/.test(label)) {
    return null;
  }

  return "subdomain can only contain lowercase letters, numbers, dashes, dots and leading underscore labels";
}

export function validateRecordValue(
  type: RecordType,
  value: string,
): string | null {
  if (!value.trim()) return "record value is required";

  if (type === "A") {
    if (!isIPv4(value)) return "invalid IPv4 address";
  }

  if (type === "AAAA") {
    if (!isIPv6(value)) return "invalid IPv6 address";
  }

  if (type === "CNAME") {
    if (value.includes("://")) return "CNAME must be a hostname, not URL";
    if (value.endsWith(".")) return "CNAME should not end with dot";
    if (!value.includes(".")) return "CNAME should be a valid hostname";
  }

  if (type === "TXT") {
    if (value.length > 255) return "TXT record is too long";
  }

  return null;
}

export function validateDomainRecord(record: DomainRecord): string | null {
  const nameError = validateSubdomain(record.name);
  if (nameError) return nameError;

  const valueError = validateRecordValue(record.type, record.value);
  if (valueError) return valueError;

  return null;
}

function isIPv4(value: string): boolean {
  const parts = value.split(".");
  if (parts.length !== 4) return false;

  return parts.every((part) => {
    if (!/^\d+$/.test(part)) return false;

    const num = Number(part);
    return num >= 0 && num <= 255;
  });
}

function isIPv6(value: string): boolean {
  return /^[a-fA-F0-9:]+$/.test(value) && value.includes(":");
}
