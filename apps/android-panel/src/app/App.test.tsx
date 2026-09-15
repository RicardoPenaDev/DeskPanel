import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import App from "./App";

describe("App", () => {
  it("renderiza a casca inicial do painel", () => {
    render(<App />);
    expect(screen.getByText("DeskPanel")).toBeInTheDocument();
  });
});
