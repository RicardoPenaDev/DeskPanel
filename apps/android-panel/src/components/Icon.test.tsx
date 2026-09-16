import { describe, expect, it } from "vitest";
import { render } from "@testing-library/react";
import Icon, { ICON_NAMES } from "./Icon";

describe("Icon", () => {
  it("renderiza um svg para cada ícone empacotado sem lançar erro", () => {
    for (const name of ICON_NAMES) {
      const { container, unmount } = render(<Icon name={name} />);
      expect(container.querySelector("svg")).toBeInTheDocument();
      unmount();
    }
  });

  it("cai no ícone genérico quando o nome é desconhecido", () => {
    const { container } = render(<Icon name="algo-que-nao-existe" />);
    expect(container.querySelector("svg")).toBeInTheDocument();
    expect(container.querySelector("rect")).toBeInTheDocument();
  });

  it("cai no ícone genérico quando nenhum nome é passado", () => {
    const { container } = render(<Icon />);
    expect(container.querySelector("svg")).toBeInTheDocument();
  });

  it("renderiza uma imagem real quando `src` é passado, em vez do desenho vetorial", () => {
    const { container } = render(<Icon name="chrome" src="blob:fake-icon" />);
    const img = container.querySelector("img");
    expect(img).toBeInTheDocument();
    expect(img).toHaveAttribute("src", "blob:fake-icon");
    expect(container.querySelector("svg")).not.toBeInTheDocument();
  });
});
