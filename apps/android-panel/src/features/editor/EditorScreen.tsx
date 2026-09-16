// Editor do painel (PROJECT.md §17 Fase 4, §10.2-C). Componente
// totalmente controlado: não guarda cópia própria do layout, só estado de
// UI efêmero (página aberta no editor, modo de reorganizar, painel de
// edição aberto). Toda mutação de dado chama `onSave` com o layout
// completo já modificado — quem persiste e é dono do estado de verdade é
// o hook `useDeskPanelConnection` (`updateLayout`/`resetLayoutToDefault`).
//
// O editor nunca inventa uma ação nova: só escolhe entre as que já vieram
// do catálogo autoritativo do Mac (`actionsCatalog`). Uma ação marcada
// como perigosa pelo Mac (`requireLongPress`) não pode ser destravada
// aqui — ver ADR-0005.

import { useState } from "react";
import type { ActionSummary } from "../../protocol/httpClient";
import type {
  ButtonColor,
  DashboardButton,
  DashboardConfig,
  DashboardPage,
} from "../../storage/layout";
import { applySlots, pageSlots } from "../../storage/pageSlots";
import Icon, { ICON_NAMES } from "../../components/Icon";
import ActionPicker from "./ActionPicker";

export interface EditorScreenProps {
  layout: DashboardConfig;
  actionsCatalog: Record<string, ActionSummary>;
  onSave: (next: DashboardConfig) => Promise<void>;
  onRestoreDefault: () => Promise<void>;
  onClose: () => void;
}

const COLORS: ButtonColor[] = ["neutral", "blue", "green", "orange", "red", "purple"];

function generateId(prefix: string): string {
  if (typeof crypto !== "undefined" && typeof crypto.randomUUID === "function") {
    return `${prefix}-${crypto.randomUUID()}`;
  }
  return `${prefix}-${Date.now()}-${Math.random().toString(16).slice(2)}`;
}

type ButtonPatch = Partial<
  Pick<DashboardButton, "labelOverride" | "iconOverride" | "color" | "requireLongPress">
>;

interface EditingSlot {
  pageId: string;
  slotIndex: number;
}

export default function EditorScreen({
  layout,
  actionsCatalog,
  onSave,
  onRestoreDefault,
  onClose,
}: EditorScreenProps) {
  const [pageIndex, setPageIndex] = useState(0);
  const [reorderMode, setReorderMode] = useState(false);
  const [selectedSlot, setSelectedSlot] = useState<number | null>(null);
  const [editingSlot, setEditingSlot] = useState<EditingSlot | null>(null);
  const [pickingAction, setPickingAction] = useState(false);
  const [confirmingReset, setConfirmingReset] = useState(false);
  const [confirmingDeletePageId, setConfirmingDeletePageId] = useState<string | null>(null);
  const [moveError, setMoveError] = useState<string | null>(null);

  const profile =
    layout.profiles.find((p) => p.id === layout.activeProfileId) ?? layout.profiles[0];
  const pages = profile?.pages ?? [];
  const safePageIndex = Math.min(pageIndex, Math.max(pages.length - 1, 0));
  const currentPage = pages[safePageIndex];

  function commitPages(nextPages: DashboardPage[]): void {
    if (!profile) return;
    const nextProfiles = layout.profiles.map((p) =>
      p.id === profile.id ? { ...p, pages: nextPages } : p,
    );
    void onSave({ ...layout, profiles: nextProfiles });
  }

  function resetTransientState(): void {
    setSelectedSlot(null);
    setEditingSlot(null);
    setPickingAction(false);
    setMoveError(null);
  }

  function handleSelectPage(index: number): void {
    setPageIndex(index);
    setConfirmingDeletePageId(null);
    resetTransientState();
  }

  function handleAddPage(): void {
    if (!profile) return;
    const newPage: DashboardPage = {
      id: generateId("page"),
      name: `Página ${profile.pages.length + 1}`,
      columns: 4,
      rows: 2,
      buttons: [],
    };
    commitPages([...profile.pages, newPage]);
    setPageIndex(profile.pages.length);
    resetTransientState();
  }

  function handleDeletePage(pageId: string): void {
    if (!profile || profile.pages.length <= 1) return;
    const nextPages = profile.pages.filter((p) => p.id !== pageId);
    commitPages(nextPages);
    setPageIndex((idx) => Math.min(idx, nextPages.length - 1));
    setConfirmingDeletePageId(null);
    resetTransientState();
  }

  function handleToggleReorder(): void {
    setReorderMode((prev) => !prev);
    resetTransientState();
  }

  function handleSlotClick(pageId: string, index: number): void {
    if (reorderMode) {
      handleReorderTap(index);
      return;
    }
    setEditingSlot({ pageId, slotIndex: index });
    setPickingAction(false);
    setMoveError(null);
  }

  function handleReorderTap(index: number): void {
    if (!currentPage || !profile) return;
    const slots = pageSlots(currentPage);

    if (selectedSlot === null) {
      if (slots[index]) setSelectedSlot(index);
      return;
    }
    if (selectedSlot === index) {
      setSelectedSlot(null);
      return;
    }

    const temp = slots[index];
    slots[index] = slots[selectedSlot];
    slots[selectedSlot] = temp;

    const nextPages = profile.pages.map((p) =>
      p.id === currentPage.id ? applySlots(p, slots) : p,
    );
    commitPages(nextPages);
    setSelectedSlot(null);
  }

  function handleChooseAction(actionId: string): void {
    if (!editingSlot || !currentPage || !profile) return;
    const catalogAction = actionsCatalog[actionId];
    const slots = pageSlots(currentPage);
    const existing = slots[editingSlot.slotIndex];

    const nextButton: DashboardButton = existing
      ? {
          ...existing,
          actionId,
          requireLongPress: catalogAction?.requireLongPress ? true : existing.requireLongPress,
        }
      : {
          id: generateId("btn"),
          actionId,
          position: editingSlot.slotIndex,
          requireLongPress: Boolean(catalogAction?.requireLongPress),
        };

    slots[editingSlot.slotIndex] = nextButton;
    const nextPages = profile.pages.map((p) =>
      p.id === editingSlot.pageId ? applySlots(p, slots) : p,
    );
    commitPages(nextPages);
    resetTransientState();
  }

  function handleUpdateButton(pageId: string, buttonId: string, patch: ButtonPatch): void {
    if (!profile) return;
    const nextPages = profile.pages.map((p) => {
      if (p.id !== pageId) return p;
      return {
        ...p,
        buttons: p.buttons.map((b) => {
          if (b.id !== buttonId) return b;
          const catalogAction = actionsCatalog[b.actionId];
          // Uma ação perigosa (kind screen_lock/display_sleep no agente)
          // nunca pode ter o toque prolongado destravado por aqui — ver
          // ADR-0005.
          const nextRequireLongPress = catalogAction?.requireLongPress
            ? true
            : (patch.requireLongPress ?? b.requireLongPress);
          return { ...b, ...patch, requireLongPress: nextRequireLongPress };
        }),
      };
    });
    commitPages(nextPages);
  }

  function handleRemoveButton(pageId: string, buttonId: string): void {
    if (!profile) return;
    const nextPages = profile.pages.map((p) =>
      p.id !== pageId ? p : { ...p, buttons: p.buttons.filter((b) => b.id !== buttonId) },
    );
    commitPages(nextPages);
    resetTransientState();
  }

  function handleMoveToPage(fromPageId: string, buttonId: string, toPageId: string): void {
    if (!profile) return;
    const fromPage = profile.pages.find((p) => p.id === fromPageId);
    const toPage = profile.pages.find((p) => p.id === toPageId);
    if (!fromPage || !toPage) {
      setMoveError("Página não encontrada");
      return;
    }
    const button = fromPage.buttons.find((b) => b.id === buttonId);
    if (!button) {
      setMoveError("Botão não encontrado");
      return;
    }

    const toSlots = pageSlots(toPage);
    const emptyIndex = toSlots.findIndex((slot) => slot === null);
    if (emptyIndex === -1) {
      setMoveError("A página de destino está cheia");
      return;
    }

    const nextPages = profile.pages.map((p) => {
      if (p.id === fromPageId) return { ...p, buttons: p.buttons.filter((b) => b.id !== buttonId) };
      if (p.id === toPageId) {
        const slots = pageSlots(p);
        slots[emptyIndex] = { ...button, position: emptyIndex };
        return applySlots(p, slots);
      }
      return p;
    });

    commitPages(nextPages);
    resetTransientState();
  }

  async function handleConfirmReset(): Promise<void> {
    await onRestoreDefault();
    setConfirmingReset(false);
    setReorderMode(false);
    setPageIndex(0);
    resetTransientState();
  }

  const editingButton =
    editingSlot && currentPage && editingSlot.pageId === currentPage.id
      ? pageSlots(currentPage)[editingSlot.slotIndex]
      : null;

  return (
    <main className="dp-editor">
      <header className="dp-editor__header">
        <h1>Editor do painel</h1>
        <button type="button" onClick={onClose}>
          Voltar ao painel
        </button>
      </header>

      <div className="dp-editor__pages">
        {pages.map((page, index) => (
          <div key={page.id} className="dp-editor__page-tab-wrap">
            <button
              type="button"
              className={`dp-editor__page-tab${index === safePageIndex ? " dp-editor__page-tab--active" : ""}`}
              onClick={() => handleSelectPage(index)}
            >
              {page.name}
            </button>
            {pages.length > 1 && (
              <button
                type="button"
                className="dp-editor__page-delete"
                aria-label={`Excluir página ${page.name}`}
                onClick={() => setConfirmingDeletePageId(page.id)}
              >
                ×
              </button>
            )}
          </div>
        ))}
        <button type="button" className="dp-editor__page-add" onClick={handleAddPage}>
          + Nova página
        </button>
      </div>

      {confirmingDeletePageId && (
        <div className="dp-editor__confirm" role="alertdialog">
          <p>Excluir esta página e os botões nela? Essa ação não pode ser desfeita.</p>
          <div className="dp-editor__confirm-actions">
            <button type="button" onClick={() => handleDeletePage(confirmingDeletePageId)}>
              Excluir
            </button>
            <button type="button" onClick={() => setConfirmingDeletePageId(null)}>
              Cancelar
            </button>
          </div>
        </div>
      )}

      <div className="dp-editor__toolbar">
        <button
          type="button"
          className={`dp-editor__reorder-toggle${reorderMode ? " dp-editor__reorder-toggle--active" : ""}`}
          onClick={handleToggleReorder}
        >
          {reorderMode ? "Concluir reorganização" : "Reorganizar botões"}
        </button>
        <button type="button" onClick={() => setConfirmingReset(true)}>
          Restaurar layout padrão
        </button>
      </div>

      {reorderMode && (
        <p className="dp-muted dp-editor__hint">
          {selectedSlot === null
            ? "Toque em um botão para selecioná-lo."
            : "Agora toque no lugar para onde ele deve ir."}
        </p>
      )}

      {confirmingReset && (
        <div className="dp-editor__confirm" role="alertdialog">
          <p>
            Restaurar o layout padrão? Todas as páginas e botões personalizados serão substituídos.
          </p>
          <div className="dp-editor__confirm-actions">
            <button type="button" onClick={() => void handleConfirmReset()}>
              Restaurar
            </button>
            <button type="button" onClick={() => setConfirmingReset(false)}>
              Cancelar
            </button>
          </div>
        </div>
      )}

      {currentPage && (
        <div
          className="dp-editor__grid"
          style={{
            gridTemplateColumns: `repeat(${currentPage.columns}, 1fr)`,
            gridTemplateRows: `repeat(${currentPage.rows}, 1fr)`,
          }}
        >
          {pageSlots(currentPage).map((button, index) => {
            const catalogAction = button ? actionsCatalog[button.actionId] : undefined;
            const selected = reorderMode && selectedSlot === index;
            const label = button
              ? (button.labelOverride ?? catalogAction?.label ?? button.actionId)
              : null;
            return (
              <button
                key={index}
                type="button"
                className={`dp-editor__slot${button ? "" : " dp-editor__slot--empty"}${
                  selected ? " dp-editor__slot--selected" : ""
                }`}
                aria-label={button ? (label ?? undefined) : "Slot vazio"}
                onClick={() => handleSlotClick(currentPage.id, index)}
              >
                {button ? (
                  <>
                    <Icon
                      name={button.iconOverride ?? catalogAction?.icon}
                      src={button.iconOverride ? undefined : catalogAction?.iconUrl}
                      className="dp-editor__slot-icon"
                    />
                    <span>{label}</span>
                  </>
                ) : (
                  <span className="dp-muted">+ adicionar</span>
                )}
              </button>
            );
          })}
        </div>
      )}

      {editingSlot && currentPage && (!editingButton || pickingAction) && (
        <div
          className="dp-editor__panel"
          role="dialog"
          aria-label={editingButton ? "Trocar ação" : "Adicionar botão"}
        >
          <ActionPicker
            actionsCatalog={actionsCatalog}
            onPick={handleChooseAction}
            onCancel={() => {
              if (editingButton) {
                setPickingAction(false);
              } else {
                resetTransientState();
              }
            }}
          />
        </div>
      )}

      {editingSlot && currentPage && editingButton && !pickingAction && (
        <div className="dp-editor__panel" role="dialog" aria-label="Editar botão">
          <div className="dp-editor__panel-header">
            <span>{actionsCatalog[editingButton.actionId]?.label ?? editingButton.actionId}</span>
            <button type="button" onClick={resetTransientState}>
              Fechar
            </button>
          </div>

          <button type="button" onClick={() => setPickingAction(true)}>
            Trocar ação
          </button>

          <label className="dp-field">
            Nome no botão
            <input
              value={editingButton.labelOverride ?? ""}
              placeholder={actionsCatalog[editingButton.actionId]?.label ?? editingButton.actionId}
              onChange={(e) =>
                handleUpdateButton(currentPage.id, editingButton.id, {
                  labelOverride: e.target.value || undefined,
                })
              }
            />
          </label>

          <div className="dp-editor__icon-grid" role="group" aria-label="Ícone">
            {ICON_NAMES.map((name) => {
              const active =
                (editingButton.iconOverride ?? actionsCatalog[editingButton.actionId]?.icon) ===
                name;
              return (
                <button
                  key={name}
                  type="button"
                  aria-label={`Ícone ${name}`}
                  aria-pressed={active}
                  className={`dp-editor__icon-option${active ? " dp-editor__icon-option--active" : ""}`}
                  onClick={() =>
                    handleUpdateButton(currentPage.id, editingButton.id, { iconOverride: name })
                  }
                >
                  <Icon name={name} />
                </button>
              );
            })}
          </div>

          <div className="dp-editor__color-grid" role="group" aria-label="Cor">
            {COLORS.map((color) => {
              const active = (editingButton.color ?? "neutral") === color;
              return (
                <button
                  key={color}
                  type="button"
                  aria-label={`Cor ${color}`}
                  aria-pressed={active}
                  className={`dp-editor__color-swatch dp-button--${color}${
                    active ? " dp-editor__color-swatch--active" : ""
                  }`}
                  onClick={() => handleUpdateButton(currentPage.id, editingButton.id, { color })}
                />
              );
            })}
          </div>

          <label className="dp-field dp-field--checkbox">
            <input
              type="checkbox"
              checked={
                Boolean(actionsCatalog[editingButton.actionId]?.requireLongPress) ||
                editingButton.requireLongPress
              }
              disabled={Boolean(actionsCatalog[editingButton.actionId]?.requireLongPress)}
              onChange={(e) =>
                handleUpdateButton(currentPage.id, editingButton.id, {
                  requireLongPress: e.target.checked,
                })
              }
            />
            Exigir toque prolongado
          </label>
          {actionsCatalog[editingButton.actionId]?.requireLongPress && (
            <p className="dp-muted">Esta ação já exige toque prolongado por segurança.</p>
          )}

          {pages.length > 1 && (
            <label className="dp-field">
              Mover para página
              <select
                value=""
                onChange={(e) => {
                  if (!e.target.value) return;
                  handleMoveToPage(currentPage.id, editingButton.id, e.target.value);
                }}
              >
                <option value="">Selecionar página…</option>
                {pages
                  .filter((p) => p.id !== currentPage.id)
                  .map((p) => (
                    <option key={p.id} value={p.id}>
                      {p.name}
                    </option>
                  ))}
              </select>
            </label>
          )}
          {moveError && (
            <p className="dp-editor__error" role="alert">
              {moveError}
            </p>
          )}

          <button
            type="button"
            className="dp-editor__remove-button"
            onClick={() => handleRemoveButton(currentPage.id, editingButton.id)}
          >
            Remover botão
          </button>
        </div>
      )}
    </main>
  );
}
