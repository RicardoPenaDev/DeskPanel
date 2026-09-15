import { describe, expect, it, vi } from "vitest";
import { render, screen } from "@testing-library/react";
import userEvent from "@testing-library/user-event";
import ActionPicker from "./ActionPicker";
import type { ActionSummary } from "../../protocol/httpClient";

function catalog(): Record<string, ActionSummary> {
  return {
    "app.chrome": {
      id: "app.chrome",
      label: "Chrome",
      icon: "chrome",
      kind: "open_app",
      requireLongPress: false,
    },
    "system.lock": {
      id: "system.lock",
      label: "Bloquear Mac",
      icon: "lock",
      kind: "screen_lock",
      requireLongPress: true,
    },
  };
}

describe("ActionPicker", () => {
  it("lista as ações do catálogo em ordem alfabética", () => {
    render(<ActionPicker actionsCatalog={catalog()} onPick={vi.fn()} onCancel={vi.fn()} />);
    const items = screen.getAllByRole("listitem").map((li) => li.textContent);
    expect(items[0]).toContain("Bloquear Mac");
    expect(items[1]).toContain("Chrome");
  });

  it("chama onPick com o actionId escolhido", async () => {
    const user = userEvent.setup();
    const onPick = vi.fn();
    render(<ActionPicker actionsCatalog={catalog()} onPick={onPick} onCancel={vi.fn()} />);

    await user.click(screen.getByRole("button", { name: /Chrome/ }));
    expect(onPick).toHaveBeenCalledWith("app.chrome");
  });

  it("mostra uma mensagem quando não há ações disponíveis", () => {
    render(<ActionPicker actionsCatalog={{}} onPick={vi.fn()} onCancel={vi.fn()} />);
    expect(screen.getByText(/Nenhuma ação disponível/)).toBeInTheDocument();
  });

  it("chama onCancel ao clicar em cancelar", async () => {
    const user = userEvent.setup();
    const onCancel = vi.fn();
    render(<ActionPicker actionsCatalog={catalog()} onPick={vi.fn()} onCancel={onCancel} />);

    await user.click(screen.getByRole("button", { name: "Cancelar" }));
    expect(onCancel).toHaveBeenCalledTimes(1);
  });
});
