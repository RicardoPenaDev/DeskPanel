// Painel principal (PROJECT.md §10.2-B): grade 4×2, troca de página por
// gesto horizontal, indicador de página, indicador de conexão no topo.
// Cada botão mostra o rótulo/ícone autoritativos vindos do catálogo do
// Mac quando disponíveis; um actionId sem correspondência aparece
// "indisponível" em vez de travar o app (§10.4). O editor (Fase 4,
// §10.2-C) abre pelo botão "Editar" no topo ou por toque prolongado em um
// slot vazio da grade.

import { useEffect, useRef, useState, type TouchEvent } from "react";
import type { DashboardConfig } from "../../storage/layout";
import { pageSlots } from "../../storage/pageSlots";
import type { ActionSummary } from "../../protocol/httpClient";
import type { ActionResult, ConnectionStatus } from "../../protocol/wsClient";
import StatusBadge from "../../components/StatusBadge";
import PageIndicator from "../../components/PageIndicator";
import DashboardButton from "../../components/DashboardButton";

export interface DashboardScreenProps {
  layout: DashboardConfig;
  actionsCatalog: Record<string, ActionSummary>;
  status: ConnectionStatus;
  macName: string | null;
  vibrationEnabled?: boolean;
  executeAction: (actionId: string) => Promise<ActionResult>;
  onOpenEditor: () => void;
  onOpenSettings: () => void;
}

const SWIPE_THRESHOLD_PX = 60;
const EMPTY_SLOT_LONG_PRESS_MS = 600;

export default function DashboardScreen({
  layout,
  actionsCatalog,
  status,
  macName,
  vibrationEnabled = true,
  executeAction,
  onOpenEditor,
  onOpenSettings,
}: DashboardScreenProps) {
  const profile =
    layout.profiles.find((p) => p.id === layout.activeProfileId) ?? layout.profiles[0];
  const pages = profile?.pages ?? [];
  const [pageIndex, setPageIndex] = useState(0);
  const touchStartX = useRef<number | null>(null);
  const emptySlotTimer = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    return () => {
      if (emptySlotTimer.current) clearTimeout(emptySlotTimer.current);
    };
  }, []);

  const safePageIndex = Math.min(pageIndex, Math.max(pages.length - 1, 0));
  const page = pages[safePageIndex];

  function handleTouchStart(event: TouchEvent): void {
    touchStartX.current = event.touches[0]?.clientX ?? null;
  }

  function handleTouchEnd(event: TouchEvent): void {
    const startX = touchStartX.current;
    touchStartX.current = null;
    if (startX === null) return;

    const endX = event.changedTouches[0]?.clientX ?? startX;
    const delta = endX - startX;

    if (delta > SWIPE_THRESHOLD_PX && safePageIndex > 0) {
      setPageIndex(safePageIndex - 1);
    } else if (delta < -SWIPE_THRESHOLD_PX && safePageIndex < pages.length - 1) {
      setPageIndex(safePageIndex + 1);
    }
  }

  async function activate(actionId: string): Promise<{ ok: boolean; message?: string }> {
    try {
      const result = await executeAction(actionId);
      if (result.status === "success") return { ok: true };
      return { ok: false, message: result.message ?? "Falha ao executar" };
    } catch (err) {
      return { ok: false, message: err instanceof Error ? err.message : "Falha ao executar" };
    }
  }

  function handleEmptySlotPointerDown(): void {
    emptySlotTimer.current = setTimeout(() => {
      emptySlotTimer.current = null;
      onOpenEditor();
    }, EMPTY_SLOT_LONG_PRESS_MS);
  }

  function clearEmptySlotTimer(): void {
    if (emptySlotTimer.current) {
      clearTimeout(emptySlotTimer.current);
      emptySlotTimer.current = null;
    }
  }

  return (
    <main className="dp-dashboard" onTouchStart={handleTouchStart} onTouchEnd={handleTouchEnd}>
      <header className="dp-dashboard__header">
        <StatusBadge status={status} macName={macName} />
        {page && <span className="dp-dashboard__page-name">{page.name}</span>}
        <PageIndicator count={pages.length} activeIndex={safePageIndex} />
        <button type="button" className="dp-dashboard__edit-button" onClick={onOpenEditor}>
          Editar
        </button>
        <button type="button" className="dp-dashboard__settings-button" onClick={onOpenSettings}>
          Configurações
        </button>
      </header>

      {page && (
        <div
          className="dp-dashboard__grid"
          style={{
            gridTemplateColumns: `repeat(${page.columns}, 1fr)`,
            gridTemplateRows: `repeat(${page.rows}, 1fr)`,
          }}
        >
          {pageSlots(page).map((button, index) => {
            if (!button) {
              return (
                <div
                  key={`empty-${index}`}
                  className="dp-dashboard__empty-slot"
                  onPointerDown={handleEmptySlotPointerDown}
                  onPointerUp={clearEmptySlotTimer}
                  onPointerLeave={clearEmptySlotTimer}
                  onPointerCancel={clearEmptySlotTimer}
                  aria-hidden="true"
                />
              );
            }

            const catalogAction = actionsCatalog[button.actionId];
            return (
              <DashboardButton
                key={button.id}
                label={button.labelOverride ?? catalogAction?.label ?? button.actionId}
                icon={button.iconOverride ?? catalogAction?.icon}
                iconUrl={button.iconOverride ? undefined : catalogAction?.iconUrl}
                color={button.color ?? "neutral"}
                requireLongPress={
                  button.requireLongPress || Boolean(catalogAction?.requireLongPress)
                }
                unavailable={!catalogAction}
                vibrationEnabled={vibrationEnabled}
                onActivate={() => activate(button.actionId)}
              />
            );
          })}
        </div>
      )}
    </main>
  );
}
