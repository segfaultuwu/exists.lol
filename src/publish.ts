import { CloudflareDNS } from "./dns/cloudflare";

const dns = new CloudflareDNS(
  process.env.CLOUDFLARE_API_TOKEN!,
  process.env.CLOUDFLARE_ZONE_ID!,
  process.env.ROOT_DOMAIN ?? "exists.lol",
);

export async function publishDomainToCloudflare(domain: {
  subdomain: string;
  type: "A" | "AAAA" | "CNAME" | "TXT" | "MX";
  value: string;
  proxied?: boolean;
  priority?: number;
}) {
  return await dns.upsertRecord({
    subdomain: domain.subdomain,
    type: domain.type,
    content: domain.value,
    proxied: domain.proxied ?? false,
    ttl: 1,
    priority: domain.priority,
  });
}
