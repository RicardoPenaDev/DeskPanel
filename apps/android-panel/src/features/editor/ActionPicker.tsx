// Lista de ações do catálogo autoritativo do Mac para escolher em um slot
// do editor (PROJECT.md §10.2-C: "escolher ação da lista recebida do
// Mac"). Nunca permite digitar um actionId livre — só o que já veio do
// agente em `actionsCatalog`.

import type { ActionSummary } from "../../protocol/httpClient";
import Icon from "../../components/Icon";

export interface ActionPickerProps {
  actionsCatalog: Record<string, ActionSummary>;
  onPick: (actionId: string) => void;
  onCancel: () => void;
}

export default function ActionPicker({ actionsCatalog, onPick, onCancel }: ActionPickerProps) {
  const actions = Object.values(actionsCatalog).sort((a, b) => a.label.localeCompare(b.label));

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
        <ul className="dp-action-picker__list">
          {actions.map((action) => (
            <li key={action.id}>
              <button type="button" onClick={() => onPick(action.id)}>
                <Icon name={action.icon} />
                <span>{action.label}</span>
              </button>
            </li>
          ))}
        </ul>
      )}
    </div>
  );
}
