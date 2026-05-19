import { defineConfig } from "vite";

export default defineConfig({
  build: {
    outDir: "web-build",
  },
  server: {
    allowedHosts: ["syncspaceedu.duskoide.org"],
  },
});
