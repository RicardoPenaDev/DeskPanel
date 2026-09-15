import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { fireEvent, render, screen, waitFor, within } from "@testing-library/react";
import DashboardScreen from "./DashboardScreen";
import { defaultDashboardConfig } from "../../storage/defaultLayout";
import type { DashboardConfig } from "../../storage/layout";
import type { ActionSummary } from "../../protocol/httpClient";

function catalogWith(ids: string[]): Record<string, ActionSummary> {
  const catalog: Record<string, ActionSummary> = {};
  for (const id of ids) {
    catalog[id] = {
      id,
      label: `Ação ${id}`,
      icon: "icon",
      kind: "open_app",
      requireLongPress: false,
    };
  }
  return catalog;
}

function grid() {
  return screen.getByRole("main").querySelector(".dp-dashboard__grid") as HTMLElement;
}

describe("DashboardScreen", () => {
  it("renderiza a grade 4x2 da primeira página com 8 botões de ação", () => {
    const layout = defaultDashboardConfig();
    render(
      <DashboardScreen
        layout={layout}
        actionsCatalog={catalogWith(layout.profiles[0].pages[0].buttons.map((b) => b.actionId))}
        status="connected"
        macName="MacBook de Ricardo"
        executeAction={vi.fn()}
        onOpenEditor={vi.fn()}
      />,
    );

    expect(within(grid()).getAllByRole("button")).toHaveLength(8);
    expect(screen.getByText("MacBook de Ricardo")).toBeInTheDocument();
  });

  it("marca como indisponível um botão cujo actionId não está no catálogo do Mac", () => {
    const layout = defaultDashboardConfig();
    render(
      <DashboardScreen
        layout={layout}
        actionsCatalog={{}}
        status="connected"
        macName="Mac"
        executeAction={vi.fn()}
        onOpenEditor={vi.fn()}
      />,
    );

    expect(screen.getAllByText("indisponível")).toHaveLength(8);
  });

  it("chama executeAction com o actionId do botão pressionado", async () => {
    const layout = defaultDashboardConfig();
    const executeAction = vi.fn(async () => ({
      actionId: "app.chrome",
      status: "success",
      durationMs: 5,
      errorCode: null,
      message: null,
    }));

    render(
      <DashboardScreen
        layout={layout}
        actionsCatalog={catalogWith(["app.chrome"])}
        status="connected"
        macName="Mac"
        executeAction={executeAction}
        onOpenEditor={vi.fn()}
      />,
    );

    const chromeButton = screen.getByRole("button", { name: "Ação app.chrome" });
    fireEvent.pointerDown(chromeButton);
    fireEvent.pointerUp(chromeButton);

    await waitFor(() => expect(executeAction).toHaveBeenCalledWith("app.chrome"));
  });

  it("troca de página com um gesto de arrastar horizontal", () => {
    const layout = defaultDashboardConfig();
    render(
      <DashboardScreen
        layout={layout}
        actionsCatalog={{}}
        status="connected"
        macName="Mac"
        executeAction={vi.fn()}
        onOpenEditor={vi.fn()}
      />,
    );

    expect(screen.getByText(layout.profiles[0].pages[0].name)).toBeInTheDocument();

    const main = screen.getByRole("main");
    fireEvent.touchStart(main, { touches: [{ clientX: 300 }] });
    fireEvent.touchEnd(main, { changedTouches: [{ clientX: 20 }] });

    expect(screen.getByText(layout.profiles[0].pages[1].name)).toBeInTheDocument();
  });

  it("ignora um arrasto curto demais para contar como troca de página", () => {
    const layout = defaultDashboardConfig();
    render(
      <DashboardScreen
        layout={layout}
        actionsCatalog={{}}
        status="connected"
        macName="Mac"
        executeAction={vi.fn()}
        onOpenEditor={vi.fn()}
      />,
    );

    const main = screen.getByRole("main");
    fireEvent.touchStart(main, { touches: [{ clientX: 300 }] });
    fireEvent.touchEnd(main, { changedTouches: [{ clientX: 280 }] });

    expect(screen.getByText(layout.profiles[0].pages[0].name)).toBeInTheDocument();
  });

  it("abre o editor ao clicar em 'Editar'", () => {
    const onOpenEditor = vi.fn();
    render(
      <DashboardScreen
        layout={defaultDashboardConfig()}
        actionsCatalog={{}}
        status="connected"
        macName="Mac"
        executeAction={vi.fn()}
        onOpenEditor={onOpenEditor}
      />,
    );

    fireEvent.click(screen.getByRole("button", { name: "Editar" }));
    expect(onOpenEditor).toHaveBeenCalledTimes(1);
  });

  describe("toque prolongado em slot vazio", () => {
    beforeEach(() => {
      vi.useFakeTimers({ shouldAdvanceTime: true });
    });

    afterEach(() => {
      vi.useRealTimers();
    });

    it("abre o editor após 600ms de toque prolongado em um slot vazio", () => {
      const onOpenEditor = vi.fn();
      const layout: DashboardConfig = defaultDashboardConfig();
      // esvazia a segunda página para ter slots vazios visíveis
      layout.profiles[0].pages[1].buttons = [];

      render(
        <DashboardScreen
          layout={layout}
          actionsCatalog={{}}
          status="connected"
          macName="Mac"
          executeAction={vi.fn()}
          onOpenEditor={onOpenEditor}
        />,
      );

      const main = screen.getByRole("main");
      fireEvent.touchStart(main, { touches: [{ clientX: 300 }] });
      fireEvent.touchEnd(main, { changedTouches: [{ clientX: 20 }] });

      const emptySlot = grid().querySelector(".dp-dashboard__empty-slot") as HTMLElement;
      fireEvent.pointerDown(emptySlot);
      vi.advanceTimersByTime(650);

      expect(onOpenEditor).toHaveBeenCalledTimes(1);
    });

    it("não abre o editor se soltar o slot vazio antes de 600ms", () => {
      const onOpenEditor = vi.fn();
      const layout: DashboardConfig = defaultDashboardConfig();
      layout.profiles[0].pages[1].buttons = [];

      render(
        <DashboardScreen
          layout={layout}
          actionsCatalog={{}}
          status="connected"
          macName="Mac"
          executeAction={vi.fn()}
          onOpenEditor={onOpenEditor}
        />,
      );

      const main = screen.getByRole("main");
      fireEvent.touchStart(main, { touches: [{ clientX: 300 }] });
      fireEvent.touchEnd(main, { changedTouches: [{ clientX: 20 }] });

      const emptySlot = grid().querySelector(".dp-dashboard__empty-slot") as HTMLElement;
      fireEvent.pointerDown(emptySlot);
      vi.advanceTimersByTime(300);
      fireEvent.pointerUp(emptySlot);
      vi.advanceTimersByTime(400);

      expect(onOpenEditor).not.toHaveBeenCalled();
    });
  });
});
