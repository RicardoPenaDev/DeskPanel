// Lista de ações do catálogo autoritativo do Mac para escolher em um slot
// do editor (PROJECT.md §10.2-C: "escolher ação da lista recebida do
// Mac"). Nunca permite digitar um actionId livre — só o que já veio do
// agente em `actionsCatalog`, que já inclui tanto as ações curadas de
// config.json quanto todo app encontrado na varredura ao vivo de
// /Applications (useDeskPanelConnection mescla as duas fontes).

import { useState } from "react";
import type { ActionSummary } from "../../protocol/httpClient";
import Icon from "../../components/Icon";

export interface ActionPickerProps {
  actionsCatalog: Record<string, ActionSummary>;
  onPick: (actionId: string) => void;
  onCancel: () => void;
}

export default function ActionPicker({ actionsCatalog, onPick, onCancel }: ActionPickerProps) {
  const [filter, setFilter] = useState("");

  const actions = Object.values(actionsCatalog).sort((a, b) => a.label.localeCompare(b.label));
  const normalizedFilter = filter.trim().toLowerCase();
  const visibleActions = normalizedFilter
    ? actions.filter((a) => a.label.toLowerCase().includes(normalizedFilter))
    : actions;

  return (
    <div className="dp-action-picker">
      <div className="dp-action-picker__header">
        <span>Escolher ação</span>
        <button type="button" onClick={onCancel}>
          Cancelar
        </button>
      </div>

      {actions.length === 0 ? (
        <p className="dp-muted">Nenhuma ação disponível — confira a conexão com o Mac.</p>
      ) : (
        <>
          {actions.length > 8 && (
            <input
              type="text"
              className="dp-action-picker__search"
              placeholder="Buscar app ou ação…"
              value={filter}
              onChange={(e) => setFilter(e.target.value)}
              aria-label="Buscar app ou ação"
            />
          )}
          {visibleActions.length === 0 ? (
            <p className="dp-muted">Nada encontrado para "{filter}".</p>
          ) : (
            <ul className="dp-action-picker__list">
              {visibleActions.map((action) => (
                <li key={action.id}>
                  <button type="button" onClick={() => onPick(action.id)}>
                    <Icon name={action.icon} src={action.iconUrl} />
                    <span>{action.label}</span>
                  </button>
                </li>
              ))}
            </ul>
          )}
        </>
      )}
    </div>
  );
}
