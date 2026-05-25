import { defineConfig } from "vite";

export default defineConfig({
  build: {
    outDir: "dist",
  },
  server: {
    allowedHosts: ["syncspaceedu.duskoide.org"],
  },
});
