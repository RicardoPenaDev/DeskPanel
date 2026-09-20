// Botão do painel com os cinco estados exigidos por PROJECT.md §10.2-B:
// normal, pressionado, executando, sucesso e erro — mais um estado
// "indisponível" para actionId sem correspondência no catálogo do Mac
// (§10.4: "instalação deve tolerar aplicativos ausentes"). Ações marcadas
// como perigosas exigem toque prolongado (§10.2-C, §7.4). Vibração curta
// em sucesso/erro é opcional, ligada por padrão, e controlada pela tela de
// Configurações (§10.2-D) via a prop vibrationEnabled.

import { useEffect, useRef, useState } from "react";
import type { ButtonColor } from "../storage/layout";
import Icon from "./Icon";
import { vibrateError, vibrateSuccess } from "../services/haptics";

export type DashboardButtonVisualState = "idle" | "pressed" | "running" | "success" | "error";

export interface ActivateOutcome {
  ok: boolean;
  message?: string;
}

export interface DashboardButtonProps {
  label: string;
  icon?: string;
  iconUrl?: string;
  color?: ButtonColor;
  requireLongPress: boolean;
  unavailable?: boolean;
  vibrationEnabled?: boolean;
  onActivate: () => Promise<ActivateOutcome>;
}

const LONG_PRESS_MS = 600;
const SUCCESS_RESET_MS = 600;
const ERROR_RESET_MS = 1800;

export default function DashboardButton({
  label,
  icon,
  iconUrl,
  color = "neutral",
  requireLongPress,
  unavailable = false,
  vibrationEnabled = true,
  onActivate,
}: DashboardButtonProps) {
  const [state, setState] = useState<DashboardButtonVisualState>("idle");
  const [errorMessage, setErrorMessage] = useState<string | null>(null);
  const longPressTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const resetTimer = useRef<ReturnType<typeof setTimeout> | null>(null);
  const longPressFired = useRef(false);

  useEffect(() => {
    return () => {
      if (longPressTimer.current) clearTimeout(longPressTimer.current);
      if (resetTimer.current) clearTimeout(resetTimer.current);
    };
  }, []);

  function clearLongPressTimer(): void {
    if (longPressTimer.current) {
      clearTimeout(longPressTimer.current);
      longPressTimer.current = null;
    }
  }

  async function activate(): Promise<void> {
    if (unavailable || state === "running") return;
    setState("running");
    setErrorMessage(null);
    try {
      const result = await onActivate();
      if (result.ok) {
        setState("success");
        if (vibrationEnabled) void vibrateSuccess();
        resetTimer.current = setTimeout(() => setState("idle"), SUCCESS_RESET_MS);
      } else {
        setErrorMessage(result.message ?? "Falha ao executar");
        setState("error");
        if (vibrationEnabled) void vibrateError();
        resetTimer.current = setTimeout(() => setState("idle"), ERROR_RESET_MS);
      }
    } catch (err) {
      setErrorMessage(err instanceof Error ? err.message : "Falha ao executar");
      setState("error");
      if (vibrationEnabled) void vibrateError();
      resetTimer.current = setTimeout(() => setState("idle"), ERROR_RESET_MS);
    }
  }

  function handlePointerDown(): void {
    if (unavailable || state === "running") return;
    longPressFired.current = false;
    setState("pressed");
    if (requireLongPress) {
      longPressTimer.current = setTimeout(() => {
        longPressFired.current = true;
        void activate();
      }, LONG_PRESS_MS);
    }
  }

  function handlePointerUp(): void {
    clearLongPressTimer();
    if (state !== "pressed") return;
    if (requireLongPress) {
      if (!longPressFired.current) setState("idle");
      return;
    }
    void activate();
  }

  function handlePointerLeaveOrCancel(): void {
    clearLongPressTimer();
    if (state === "pressed") setState("idle");
  }

  return (
    <button
      type="button"
      className={`dp-button dp-button--${color} dp-button--${state}`}
      disabled={unavailable}
      onPointerDown={handlePointerDown}
      onPointerUp={handlePointerUp}
      onPointerLeave={handlePointerLeaveOrCancel}
      onPointerCancel={handlePointerLeaveOrCancel}
      aria-label={label}
      aria-disabled={unavailable}
    >
      <Icon name={icon} src={iconUrl} className="dp-button__icon" />
      <span className="dp-button__label">{label}</span>
      {requireLongPress && state === "idle" && <span className="dp-button__hint">segure</span>}
      {state === "success" && (
        <span className="dp-button__feedback dp-button__feedback--success" role="status">
          ✓
        </span>
      )}
      {state === "running" && (
        <span className="dp-button__spinner" role="status" aria-label="Executando">
          …
        </span>
      )}
      {state === "error" && errorMessage && (
        <span className="dp-button__error" role="alert">
          <span className="dp-button__feedback dp-button__feedback--error" aria-hidden="true">
            !
          </span>
          {errorMessage}
        </span>
      )}
      {unavailable && <span className="dp-button__unavailable">indisponível</span>}
    </button>
  );
}
