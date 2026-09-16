import { describe, expect, it } from "vitest";
import { DeviceControlWeb } from "./deviceControlWeb";

describe("DeviceControlWeb (fallback de desenvolvimento)", () => {
  it("setKeepAwake não lança fora do Android", async () => {
    const control = new DeviceControlWeb();
    await expect(control.setKeepAwake({ enabled: true })).resolves.toBeUndefined();
  });

  it("setImmersiveMode não lança fora do Android", async () => {
    const control = new DeviceControlWeb();
    await expect(control.setImmersiveMode({ enabled: true })).resolves.toBeUndefined();
  });

  it("setDimBrightness não lança fora do Android", async () => {
    const control = new DeviceControlWeb();
    await expect(control.setDimBrightness({ enabled: true })).resolves.toBeUndefined();
  });
});
