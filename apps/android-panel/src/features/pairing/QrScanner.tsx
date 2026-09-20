import { BrowserQRCodeReader } from "@zxing/browser";
import { useEffect, useRef, useState } from "react";

interface QrScannerProps {
  onCode: (value: string) => void;
  onClose: () => void;
}

export default function QrScanner({ onCode, onClose }: QrScannerProps) {
  const videoRef = useRef<HTMLVideoElement>(null);
  const controlsRef = useRef<{ stop: () => void } | null>(null);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const reader = new BrowserQRCodeReader();
    const video = videoRef.current;
    if (!video) return;
    let active = true;
    void reader.decodeFromVideoDevice(undefined, video, (result, error) => {
      if (!active) return;
      if (result) {
        active = false;
        controlsRef.current?.stop();
        onCode(result.getText());
        return;
      }
      if (error && error.name !== "NotFoundException") setError("Não foi possível acessar a câmera.");
    }).then((controls) => {
      controlsRef.current = controls;
    }).catch(() => setError("Permita o acesso à câmera para escanear o QR Code."));

    return () => {
      active = false;
      controlsRef.current?.stop();
    };
  }, [onCode]);

  return (
    <div className="dp-qr-scanner" role="dialog" aria-label="Escanear QR Code">
      <div className="dp-qr-scanner__viewport">
        <video ref={videoRef} className="dp-qr-scanner__video" muted playsInline />
        <div className="dp-qr-scanner__guide" aria-hidden="true" />
      </div>
      <p>{error ?? "Aponte a câmera para o QR Code do Mac."}</p>
      <button type="button" onClick={onClose}>Cancelar</button>
    </div>
  );
}
