import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import App from "./App";

describe("App", () => {
  it("mostra a tela de carregamento antes da configuração local terminar de carregar", () => {
    render(<App />);
    expect(screen.getByText("DeskPanel")).toBeInTheDocument();
    expect(screen.getByText("Carregando…")).toBeInTheDocument();
  });
});
