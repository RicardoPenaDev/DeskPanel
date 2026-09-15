import { describe, expect, it } from "vitest";
import { applySlots, pageSlots } from "./pageSlots";
import type { DashboardPage } from "./layout";

function makePage(): DashboardPage {
  return {
    id: "page-1",
    name: "Página 1",
    columns: 4,
    rows: 2,
    buttons: [
      { id: "b1", actionId: "app.chrome", position: 0, requireLongPress: false },
      { id: "b2", actionId: "app.whatsapp", position: 3, requireLongPress: false },
    ],
  };
}

describe("pageSlots/applySlots", () => {
  it("cria um array de 8 posições com os botões no lugar certo e o resto null", () => {
    const slots = pageSlots(makePage());
    expect(slots).toHaveLength(8);
    expect(slots[0]?.actionId).toBe("app.chrome");
    expect(slots[3]?.actionId).toBe("app.whatsapp");
    expect(slots[1]).toBeNull();
  });

  it("ignora um botão com position fora da grade em vez de travar", () => {
    const page = makePage();
    page.buttons.push({
      id: "b3",
      actionId: "shortcut.work",
      position: 99,
      requireLongPress: false,
    });
    const slots = pageSlots(page);
    expect(slots).toHaveLength(8);
    expect(slots.some((s) => s?.id === "b3")).toBe(false);
  });

  it("applySlots recalcula position a partir do índice e remove nulls", () => {
    const page = makePage();
    const slots = pageSlots(page);
    // move b1 (estava no slot 0) para o slot 5
    slots[5] = slots[0];
    slots[0] = null;

    const result = applySlots(page, slots);
    const b1 = result.buttons.find((b) => b.id === "b1");
    expect(b1?.position).toBe(5);
    expect(result.buttons).toHaveLength(2);
  });

  it("é reversível: pageSlots(applySlots(page, slots)) preserva o conteúdo", () => {
    const page = makePage();
    const slots = pageSlots(page);
    const roundTripped = pageSlots(applySlots(page, slots));
    expect(roundTripped.map((b) => b?.id ?? null)).toEqual(slots.map((b) => b?.id ?? null));
  });
});
