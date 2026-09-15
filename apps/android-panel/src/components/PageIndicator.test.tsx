import { describe, expect, it } from "vitest";
import { render, screen } from "@testing-library/react";
import PageIndicator from "./PageIndicator";

describe("PageIndicator", () => {
  it("renderiza um marcador por página e destaca a página ativa", () => {
    render(<PageIndicator count={2} activeIndex={1} />);
    const tabs = screen.getAllByRole("tab");
    expect(tabs).toHaveLength(2);
    expect(tabs[0]).toHaveAttribute("aria-selected", "false");
    expect(tabs[1]).toHaveAttribute("aria-selected", "true");
  });

  it("não renderiza nada quando há uma página só", () => {
    const { container } = render(<PageIndicator count={1} activeIndex={0} />);
    expect(container).toBeEmptyDOMElement();
  });
});
