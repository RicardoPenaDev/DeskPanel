import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import StatusBadge from "./StatusBadge";

describe("StatusBadge", () => {
  it("mostra o nome do Mac quando conectado", () => {
    render(<StatusBadge status="connected" macName="MacBook de Ricardo" />);
    expect(screen.getByText("MacBook de Ricardo")).toBeInTheDocument();
  });

  it("mostra um rótulo genérico quando reconectando, sem nome do Mac", () => {
    render(<StatusBadge status="reconnecting" macName={null} />);
    expect(screen.getByText("Reconectando…")).toBeInTheDocument();
  });

  it("mostra 'Desconectado' no estado inicial", () => {
    render(<StatusBadge status="disconnected" />);
    expect(screen.getByText("Desconectado")).toBeInTheDocument();
  });
});
