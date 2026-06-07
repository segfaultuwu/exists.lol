export const config = {
  port: Number(process.env.API_PORT ?? 3000),
  host: process.env.API_HOST ?? "localhost",
  registryDir: process.env.REGISTRY_DIR ?? "domains",
  baseDomain: process.env.BASE_DOMAIN ?? "exists.lol",
  adminToken: process.env.API_TOKEN ?? "dev-secret",

  cloudflare: {
    apiToken: process.env.CLOUDFLARE_API_TOKEN ?? "",
    zoneId: process.env.CLOUDFLARE_ZONE_ID ?? "",
  },
};
