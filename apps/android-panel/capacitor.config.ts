import type { CapacitorConfig } from "@capacitor/cli";

// PROJECT.md §5.1/§10: painel local-first, sem servidor externo. O IP/porta
// do Mac é configurado dentro do app (tela de pareamento), não aqui.
const config: CapacitorConfig = {
  appId: "dev.ricardopena.deskpanel",
  appName: "DeskPanel",
  webDir: "dist",
  // PROJECT.md §11.1: tráfego HTTP claro só para a rede local do MVP.
  // androidScheme "http" é necessário além de cleartext: true — sem isso o
  // app é servido em https://localhost e o WebView (Chromium) bloqueia
  // qualquer fetch() para http://<ip-lan> como "Mixed Content", já que
  // cleartext/usesCleartextTraffic só controla a política de rede nativa
  // do Android, não a política de mixed content do próprio WebView.
  server: {
    cleartext: true,
    androidScheme: "http",
  },
};

export default config;
