// Indicador discreto da página atual (PROJECT.md §10.2-B).

export interface PageIndicatorProps {
  count: number;
  activeIndex: number;
}

export default function PageIndicator({ count, activeIndex }: PageIndicatorProps) {
  if (count <= 1) return null;

  return (
    <div className="dp-page-indicator" role="tablist" aria-label="Páginas do painel">
      {Array.from({ length: count }, (_, index) => (
        <span
          key={index}
          role="tab"
          aria-selected={index === activeIndex}
          className={`dp-page-dot${index === activeIndex ? " dp-page-dot--active" : ""}`}
        />
      ))}
    </div>
  );
}
