import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, render, screen } from "@testing-library/react";
import FlipClock from "./FlipClock";

describe("FlipClock", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date(2026, 8, 16, 9, 5, 7));
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("mostra hora, minuto e segundo atuais como dígitos separados", () => {
    render(<FlipClock />);
    // 09:05:07 -> dígitos 0 9 0 5 0 7 (podem se repetir na tela).
    expect(screen.getAllByText("0").length).toBeGreaterThan(0);
    expect(screen.getAllByText("9").length).toBeGreaterThan(0);
    expect(screen.getAllByText("5").length).toBeGreaterThan(0);
    expect(screen.getAllByText("7").length).toBeGreaterThan(0);
  });

  it("mostra a data por baixo do relógio", () => {
    render(<FlipClock />);
    expect(screen.getByText(/2026/)).toBeInTheDocument();
  });

  it("atualiza o segundo a cada 1s sem lançar erro", () => {
    render(<FlipClock />);
    act(() => {
      vi.advanceTimersByTime(1000);
    });
    expect(screen.getAllByText("8").length).toBeGreaterThan(0);
  });

  it("não mostra bloco de clima quando não informado", () => {
    const { container } = render(<FlipClock />);
    expect(container.querySelector(".dp-flip-clock__weather")).not.toBeInTheDocument();
  });

  it("mostra o clima quando informado", () => {
    render(<FlipClock weather={{ city: "São Paulo", tempC: 24.4, description: "céu limpo" }} />);
    expect(screen.getByText("24°")).toBeInTheDocument();
    expect(screen.getByText(/céu limpo/i)).toBeInTheDocument();
    expect(screen.getByText(/São Paulo/)).toBeInTheDocument();
  });
});
