// Pacote de ícones locais (PROJECT.md §10.1: "ícones empacotados
// localmente", §17 Fase 4: "trocar ícone dentre os ícones empacotados").
// Formas simples desenhadas à mão em SVG inline — sem depender de nenhuma
// biblioteca externa de ícones. As chaves espelham o campo `icon` usado em
// configs/config.example.json; um nome desconhecido cai no ícone genérico
// em vez de quebrar o botão.

import type { SVGProps } from "react";

export const ICON_NAMES = [
  "app",
  "chrome",
  "message-circle",
  "folder",
  "terminal",
  "music",
  "code",
  "camera",
  "briefcase-business",
  "skip-back",
  "play-pause",
  "skip-forward",
  "volume-x",
  "volume-1",
  "volume-2",
  "monitor-off",
  "lock",
] as const;

export type IconName = (typeof ICON_NAMES)[number];

export interface IconProps extends Omit<SVGProps<SVGSVGElement>, "name"> {
  name?: string;
  // Ícone real do app (PNG convertido do .icns pelo Mac — ver
  // fetchAppIcon), como na tela inicial do iPhone. Tem prioridade sobre
  // `name`; ausente ou falho, cai no desenho vetorial.
  src?: string;
  size?: number;
}

function Shape({ name }: { name: IconName }) {
  switch (name) {
    case "chrome":
      return (
        <>
          <circle cx="12" cy="12" r="8" />
          <circle cx="12" cy="12" r="2.5" />
        </>
      );
    case "message-circle":
      return <path d="M4 5h16v10H9l-4 4V5Z" />;
    case "folder":
      return <path d="M3 6h6l2 2h10v10H3V6Z" />;
    case "terminal":
      return (
        <>
          <rect x="3" y="4" width="18" height="16" rx="1.5" />
          <path d="M7 9l3 3-3 3M12 15h5" />
        </>
      );
    case "music":
      return (
        <>
          <circle cx="7" cy="18" r="2.5" />
          <circle cx="17" cy="16" r="2.5" />
          <path d="M9.5 18V6l10-2v12" />
        </>
      );
    case "code":
      return <path d="M9 7 4 12l5 5M15 7l5 5-5 5" />;
    case "camera":
      return (
        <>
          <path d="M4 8h3l2-2h6l2 2h3v11H4V8Z" />
          <circle cx="12" cy="13.5" r="3.5" />
        </>
      );
    case "briefcase-business":
      return (
        <>
          <rect x="3" y="8" width="18" height="12" rx="1.5" />
          <path d="M8 8V6a2 2 0 0 1 2-2h4a2 2 0 0 1 2 2v2" />
        </>
      );
    case "skip-back":
      return <path d="M18 6v12l-9-6 9-6ZM7 6v12" />;
    case "play-pause":
      return <path d="M6 5v14l11-7-11-7Z" />;
    case "skip-forward":
      return <path d="M6 6v12l9-6-9-6ZM17 6v12" />;
    case "volume-x":
      return <path d="M4 9v6h4l5 4V5L8 9H4ZM16 9l5 6M21 9l-5 6" />;
    case "volume-1":
      return <path d="M4 9v6h4l5 4V5L8 9H4ZM16 10a3 3 0 0 1 0 4" />;
    case "volume-2":
      return <path d="M4 9v6h4l5 4V5L8 9H4ZM15.5 9a5 5 0 0 1 0 6M18.5 6.5a9 9 0 0 1 0 11" />;
    case "monitor-off":
      return (
        <>
          <rect x="3" y="4" width="18" height="12" rx="1.5" />
          <path d="M8 20h8M12 16v4M3 3l18 18" />
        </>
      );
    case "lock":
      return (
        <>
          <rect x="5" y="11" width="14" height="9" rx="1.5" />
          <path d="M8 11V8a4 4 0 0 1 8 0v3" />
        </>
      );
    case "app":
    default:
      return <rect x="4" y="4" width="16" height="16" rx="3" />;
  }
}

/**
 * Ícone monocromático via `currentColor` — herda a cor do texto do botão,
 * então funciona nos temas claro/escuro sem variante própria.
 */
export default function Icon({ name, src, size = 22, className, ...svgProps }: IconProps) {
  if (src) {
    return (
      <img
        src={src}
        width={size}
        height={size}
        alt=""
        aria-hidden="true"
        draggable={false}
        className={["dp-icon-image", className].filter(Boolean).join(" ")}
      />
    );
  }

  const resolved: IconName = (ICON_NAMES as readonly string[]).includes(name ?? "")
    ? (name as IconName)
    : "app";

  return (
    <svg
      width={size}
      height={size}
      viewBox="0 0 24 24"
      fill="none"
      stroke="currentColor"
      strokeWidth={1.6}
      strokeLinecap="round"
      strokeLinejoin="round"
      aria-hidden="true"
      className={className}
      {...svgProps}
    >
      <Shape name={resolved} />
    </svg>
  );
}
