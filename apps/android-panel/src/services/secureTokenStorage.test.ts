import { afterEach, describe, expect, it, vi } from "vitest";

const fakePlugin = vi.hoisted(() => ({
  setToken: vi.fn(async () => {}),
  getToken: vi.fn(async () => ({ token: null as string | null })),
  clearToken: vi.fn(async () => {}),
}));

vi.mock("@capacitor/core", async () => {
  const actual = await vi.importActual<typeof import("@capacitor/core")>("@capacitor/core");
  return {
    ...actual,
    registerPlugin: () => fakePlugin,
  };
});

import { clearAccessToken, readAccessToken, saveAccessToken } from "./secureTokenStorage";

describe("secureTokenStorage (wrapper do plugin nativo)", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("saveAccessToken delega para setToken do plugin com deviceId e token", async () => {
    await saveAccessToken("device-1", "tok-secreto");
    expect(fakePlugin.setToken).toHaveBeenCalledWith({
      deviceId: "device-1",
      token: "tok-secreto",
    });
  });

  it("readAccessToken devolve o token retornado pelo plugin", async () => {
    fakePlugin.getToken.mockResolvedValueOnce({ token: "tok-lido" });
    const token = await readAccessToken("device-1");
    expect(token).toBe("tok-lido");
    expect(fakePlugin.getToken).toHaveBeenCalledWith({ deviceId: "device-1" });
  });

  it("readAccessToken devolve null quando o plugin não tem token salvo", async () => {
    fakePlugin.getToken.mockResolvedValueOnce({ token: null });
    const token = await readAccessToken("device-1");
    expect(token).toBeNull();
  });

  it("clearAccessToken delega para clearToken do plugin", async () => {
    await clearAccessToken("device-1");
    expect(fakePlugin.clearToken).toHaveBeenCalledWith({ deviceId: "device-1" });
  });
});
