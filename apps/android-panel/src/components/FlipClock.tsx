// Relógio "ambiente" estilo flip-clock (dígitos em cartão preto/branco
// que viram ao trocar) — tela revelada ao arrastar para a direita a
// partir da primeira página do painel (ver DashboardScreen). Puramente
// decorativo: não depende de nenhuma fonte/asset externo, só CSS.

import { useEffect, useRef, useState } from "react";

const FLIP_DURATION_MS = 380;

// Só a primeira letra maiúscula ("Céu limpo", "Quarta-feira") — o
// text-transform: capitalize do CSS deixaria cada palavra maiúscula
// ("Céu Limpo", "Quarta-Feira"), errado em português.
function capitalizeFirst(text: string): string {
  return text.length > 0 ? text[0].toUpperCase() + text.slice(1) : text;
}

interface FlipUnitProps {
  value: string;
}

function FlipUnit({ value }: FlipUnitProps) {
  const [displayValue, setDisplayValue] = useState(value);
  const [previousValue, setPreviousValue] = useState(value);
  const [flipping, setFlipping] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (value === displayValue) return;
    setPreviousValue(displayValue);
    setDisplayValue(value);
    setFlipping(true);
    timerRef.current = setTimeout(() => setFlipping(false), FLIP_DURATION_MS);
    return () => {
      if (timerRef.current) clearTimeout(timerRef.current);
    };
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, [value]);

  return (
    <div className="dp-flip-unit">
      <div className="dp-flip-unit__card dp-flip-unit__card--top">
        <span>{displayValue}</span>
      </div>
      <div className="dp-flip-unit__card dp-flip-unit__card--bottom">
        <span>{displayValue}</span>
      </div>
      {flipping && (
        <>
          <div className="dp-flip-unit__card dp-flip-unit__card--top dp-flip-unit__flap dp-flip-unit__flap--top">
            <span>{previousValue}</span>
          </div>
          <div className="dp-flip-unit__card dp-flip-unit__card--bottom dp-flip-unit__flap dp-flip-unit__flap--bottom">
            <span>{displayValue}</span>
          </div>
        </>
      )}
      <div className="dp-flip-unit__hinge" />
    </div>
  );
}

export interface WeatherSummary {
  city: string;
  tempC: number;
  description: string;
}

export interface FlipClockProps {
  weather?: WeatherSummary | null;
}

export default function FlipClock({ weather }: FlipClockProps) {
  const [now, setNow] = useState(() => new Date());

  useEffect(() => {
    const id = setInterval(() => setNow(new Date()), 1000);
    return () => clearInterval(id);
  }, []);

  const hh = String(now.getHours()).padStart(2, "0");
  const mm = String(now.getMinutes()).padStart(2, "0");
  const ss = String(now.getSeconds()).padStart(2, "0");
  const dateLabel = capitalizeFirst(
    now.toLocaleDateString("pt-BR", {
      weekday: "long",
      day: "2-digit",
      month: "long",
      year: "numeric",
    }),
  );

  return (
    <div className="dp-flip-clock">
      {weather && (
        <div className="dp-flip-clock__weather">
          <span className="dp-flip-clock__weather-temp">{Math.round(weather.tempC)}°</span>
          <span className="dp-flip-clock__weather-desc">
            {capitalizeFirst(weather.description)} · {weather.city}
          </span>
        </div>
      )}

      <div className="dp-flip-clock__row">
        <FlipUnit value={hh[0]} />
        <FlipUnit value={hh[1]} />
        <span className="dp-flip-clock__colon">:</span>
        <FlipUnit value={mm[0]} />
        <FlipUnit value={mm[1]} />
        <span className="dp-flip-clock__colon">:</span>
        <FlipUnit value={ss[0]} />
        <FlipUnit value={ss[1]} />
      </div>

      <div className="dp-flip-clock__date">{dateLabel}</div>
    </div>
  );
}
