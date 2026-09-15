import { defineConfig } from "vitest/config";
import react from "@vitejs/plugin-react";

// DeskPanel Android panel — dev server config. Production builds são
// empacotados no APK pelo Capacitor (capacitor.config.ts); o output aqui só
// precisa ser um bundle estático que o Capacitor copia para dentro do app.
export default defineConfig({
  plugins: [react()],
  server: {
    port: 5173,
  },
  test: {
    environment: "jsdom",
    globals: true,
    setupFiles: ["./src/test/setup.ts"],
  },
});
