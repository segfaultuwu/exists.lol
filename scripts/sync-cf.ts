const apiUrl = process.env.API_URL ?? "https://api.exists.lol";
const adminToken = process.env.ADMIN_TOKEN;

if (!adminToken) {
  throw new Error("Missing ADMIN_TOKEN");
}

const res = await fetch(`${apiUrl}/api/sync/cloudflare`, {
  method: "POST",
  headers: {
    Authorization: `Bearer ${adminToken}`,
  },
});

const data = await res.json();

console.log(JSON.stringify(data, null, 2));

if (!res.ok || !(data as any).ok) {
  process.exit(1);
}
