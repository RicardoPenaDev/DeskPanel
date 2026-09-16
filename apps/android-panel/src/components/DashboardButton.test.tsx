import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { act, fireEvent, render, screen, waitFor } from "@testing-library/react";
import DashboardButton from "./DashboardButton";

const haptics = vi.hoisted(() => ({
  vibrateSuccess: vi.fn(async () => {}),
  vibrateError: vi.fn(async () => {}),
}));

vi.mock("../services/haptics", () => haptics);

describe("DashboardButton", () => {
  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
  });

  afterEach(() => {
    vi.useRealTimers();
    vi.clearAllMocks();
  });

  it("dispara a ação com um toque simples quando não exige toque prolongado", async () => {
    const onActivate = vi.fn(async () => ({ ok: true }));
    render(<DashboardButton label="Chrome" requireLongPress={false} onActivate={onActivate} />);

    const button = screen.getByRole("button", { name: "Chrome" });
    fireEvent.pointerDown(button);
    fireEvent.pointerUp(button);

    await waitFor(() => expect(onActivate).toHaveBeenCalledTimes(1));
  });

  it("mostra estado de sucesso e depois volta para o normal", async () => {
    const onActivate = vi.fn(async () => ({ ok: true }));
    render(<DashboardButton label="Chrome" requireLongPress={false} onActivate={onActivate} />);

    const button = screen.getByRole("button", { name: "Chrome" });
    fireEvent.pointerDown(button);
    fireEvent.pointerUp(button);

    await waitFor(() => expect(button.className).toContain("dp-button--success"));

    act(() => {
      vi.advanceTimersByTime(700);
    });
    await waitFor(() => expect(button.className).toContain("dp-button--idle"));
  });

  it("mostra a mensagem de erro quando a ação falha", async () => {
    const onActivate = vi.fn(async () => ({ ok: false, message: "Ação não permitida" }));
    render(
      <DashboardButton label="Bloquear Mac" requireLongPress={false} onActivate={onActivate} />,
    );

    const button = screen.getByRole("button", { name: "Bloquear Mac" });
    fireEvent.pointerDown(button);
    fireEvent.pointerUp(button);

    await waitFor(() => expect(screen.getByText("Ação não permitida")).toBeInTheDocument());
  });

  it("não dispara uma ação perigosa com um toque curto e dispara com toque prolongado", async () => {
    const onActivate = vi.fn(async () => ({ ok: true }));
    render(<DashboardButton label="Bloquear Mac" requireLongPress onActivate={onActivate} />);

    const button = screen.getByRole("button", { name: "Bloquear Mac" });

    fireEvent.pointerDown(button);
    fireEvent.pointerUp(button);
    expect(onActivate).not.toHaveBeenCalled();

    fireEvent.pointerDown(button);
    act(() => {
      vi.advanceTimersByTime(650);
    });
    await waitFor(() => expect(onActivate).toHaveBeenCalledTimes(1));
    fireEvent.pointerUp(button);
  });

  it("não dispara nada quando o botão está indisponível", () => {
    const onActivate = vi.fn(async () => ({ ok: true }));
    render(
      <DashboardButton
        label="Ação desconhecida"
        requireLongPress={false}
        unavailable
        onActivate={onActivate}
      />,
    );

    const button = screen.getByRole("button", { name: "Ação desconhecida" });
    expect(button).toBeDisabled();
    expect(screen.getByText("indisponível")).toBeInTheDocument();

    fireEvent.pointerDown(button);
    fireEvent.pointerUp(button);
    expect(onActivate).not.toHaveBeenCalled();
  });

  it("vibra em sucesso por padrão (vibrationEnabled implícito)", async () => {
    const onActivate = vi.fn(async () => ({ ok: true }));
    render(<DashboardButton label="Chrome" requireLongPress={false} onActivate={onActivate} />);

    const button = screen.getByRole("button", { name: "Chrome" });
    fireEvent.pointerDown(button);
    fireEvent.pointerUp(button);

    await waitFor(() => expect(haptics.vibrateSuccess).toHaveBeenCalledTimes(1));
    expect(haptics.vibrateError).not.toHaveBeenCalled();
  });

  it("vibra em erro quando a ação falha", async () => {
    const onActivate = vi.fn(async () => ({ ok: false, message: "Ação não permitida" }));
    render(
      <DashboardButton label="Bloquear Mac" requireLongPress={false} onActivate={onActivate} />,
    );

    const button = screen.getByRole("button", { name: "Bloquear Mac" });
    fireEvent.pointerDown(button);
    fireEvent.pointerUp(button);

    await waitFor(() => expect(haptics.vibrateError).toHaveBeenCalledTimes(1));
    expect(haptics.vibrateSuccess).not.toHaveBeenCalled();
  });

  it("não vibra quando vibrationEnabled=false", async () => {
    const onActivate = vi.fn(async () => ({ ok: true }));
    render(
      <DashboardButton
        label="Chrome"
        requireLongPress={false}
        vibrationEnabled={false}
        onActivate={onActivate}
      />,
    );

    const button = screen.getByRole("button", { name: "Chrome" });
    fireEvent.pointerDown(button);
    fireEvent.pointerUp(button);

    await waitFor(() => expect(onActivate).toHaveBeenCalledTimes(1));
    expect(haptics.vibrateSuccess).not.toHaveBeenCalled();
    expect(haptics.vibrateError).not.toHaveBeenCalled();
  });
});
