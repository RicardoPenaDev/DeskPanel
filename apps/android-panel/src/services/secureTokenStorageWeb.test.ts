import { describe, expect, it } from "vitest";
import { SecureTokenStorageWeb } from "./secureTokenStorageWeb";

describe("SecureTokenStorageWeb (fallback de desenvolvimento)", () => {
  it("guarda e devolve o token do dispositivo", async () => {
    const storage = new SecureTokenStorageWeb();
    await storage.setToken({ deviceId: "device-1", token: "tok-123" });

    const { token } = await storage.getToken({ deviceId: "device-1" });
    expect(token).toBe("tok-123");
  });

  it("retorna null para um dispositivo sem token salvo", async () => {
    const storage = new SecureTokenStorageWeb();
    const { token } = await storage.getToken({ deviceId: "desconhecido" });
    expect(token).toBeNull();
  });

  it("remove o token ao limpar", async () => {
    const storage = new SecureTokenStorageWeb();
    await storage.setToken({ deviceId: "device-1", token: "tok-123" });
    await storage.clearToken({ deviceId: "device-1" });

    const { token } = await storage.getToken({ deviceId: "device-1" });
    expect(token).toBeNull();
  });

  it("mantém tokens de dispositivos diferentes isolados", async () => {
    const storage = new SecureTokenStorageWeb();
    await storage.setToken({ deviceId: "device-1", token: "tok-1" });
    await storage.setToken({ deviceId: "device-2", token: "tok-2" });

    expect((await storage.getToken({ deviceId: "device-1" })).token).toBe("tok-1");
    expect((await storage.getToken({ deviceId: "device-2" })).token).toBe("tok-2");
  });
});
