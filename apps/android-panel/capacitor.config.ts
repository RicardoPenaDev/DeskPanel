import type { CapacitorConfig } from "@capacitor/cli";

// PROJECT.md §5.1/§10: painel local-first, sem servidor externo. O IP/porta
// do Mac é configurado dentro do app (tela de pareamento), não aqui.
const config: CapacitorConfig = {
  appId: "dev.ricardopena.deskpanel",
  appName: "DeskPanel",
  webDir: "dist",
  // PROJECT.md §11.1: tráfego HTTP claro só para a rede local do MVP.
  server: {
    cleartext: true,
  },
};

export default config;
