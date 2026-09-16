import { afterEach, describe, expect, it, vi } from "vitest";

const fakePlugin = vi.hoisted(() => ({
  setKeepAwake: vi.fn(async () => {}),
  setImmersiveMode: vi.fn(async () => {}),
  setDimBrightness: vi.fn(async () => {}),
}));

vi.mock("@capacitor/core", async () => {
  const actual = await vi.importActual<typeof import("@capacitor/core")>("@capacitor/core");
  return {
    ...actual,
    registerPlugin: () => fakePlugin,
  };
});

import { setDimBrightness, setImmersiveMode, setKeepAwake } from "./deviceControl";

describe("deviceControl (wrapper do plugin nativo)", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("setKeepAwake delega para o plugin com enabled=true", async () => {
    await setKeepAwake(true);
    expect(fakePlugin.setKeepAwake).toHaveBeenCalledWith({ enabled: true });
  });

  it("setKeepAwake delega para o plugin com enabled=false", async () => {
    await setKeepAwake(false);
    expect(fakePlugin.setKeepAwake).toHaveBeenCalledWith({ enabled: false });
  });

  it("setImmersiveMode delega para o plugin", async () => {
    await setImmersiveMode(true);
    expect(fakePlugin.setImmersiveMode).toHaveBeenCalledWith({ enabled: true });
  });

  it("setDimBrightness delega para o plugin", async () => {
    await setDimBrightness(true);
    expect(fakePlugin.setDimBrightness).toHaveBeenCalledWith({ enabled: true });
  });
});
