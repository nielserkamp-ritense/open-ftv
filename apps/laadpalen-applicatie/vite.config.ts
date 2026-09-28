import { defineConfig, type Plugin } from "vite";
import react from "@vitejs/plugin-react-swc";
import path from "path";
import { writeFileSync } from "fs";

// Vite empties web/dist on every build; the committed .gitkeep keeps the Go embed compiling on a clean checkout.
const keepEmbedPlaceholder: Plugin = {
  name: "keep-embed-placeholder",
  apply: "build",
  closeBundle() {
    writeFileSync(path.resolve(__dirname, "web/dist/.gitkeep"), "");
  },
};

// https://vitejs.dev/config/
export default defineConfig({
  server: {
    host: "::",
    port: 8080,
  },
  plugins: [react(), keepEmbedPlaceholder],
  resolve: {
    alias: {
      "@": path.resolve(__dirname, "./src"),
    },
  },
  build: {
    // The Go web server embeds the build output from web/dist.
    outDir: "web/dist",
  },
});
