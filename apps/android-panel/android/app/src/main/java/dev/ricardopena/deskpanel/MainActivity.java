package dev.ricardopena.deskpanel;

import android.os.Build;
import android.os.Bundle;
import android.view.WindowManager;

import androidx.core.view.WindowCompat;

import com.getcapacitor.BridgeActivity;

/**
 * Ponto de entrada nativo do DeskPanel. Responsabilidades adicionadas
 * à casca gerada pelo Capacitor (PROJECT.md §10.1):
 *
 *  - registrar o plugin local SecureTokenStoragePlugin (token no Android
 *    Keystore, nunca em Preferences ou arquivo);
 *  - registrar o plugin local DeviceControlPlugin (keep awake, modo
 *    imersivo e brilho reduzido, todos ajustáveis pela tela de
 *    Configurações — PROJECT.md §10.2-D);
 *  - manter a tela ativa por padrão assim que a Activity é criada ("keep
 *    awake"), antes mesmo do JS carregar a preferência salva — a tela de
 *    Configurações pode desligar isso depois via DeviceControlPlugin;
 *  - permitir que o conteúdo ocupe a área do recorte de câmera
 *    (LAYOUT_IN_DISPLAY_CUTOUT_MODE_SHORT_EDGES/ALWAYS) — sem isso o
 *    Android reserva ali uma faixa sólida do lado de fora da WebView
 *    mesmo com o modo imersivo ligado, já que "imersivo" só esconde as
 *    barras de sistema, não redesenha o recorte físico da câmera;
 *  - desligar decorFitsSystemWindows (WindowCompat) para a WebView usar
 *    a tela inteira: sem isso o Android ainda reserva, como padding do
 *    próprio layout raiz, o espaço da barra de status/recorte mesmo com
 *    a janela liberada para desenhar ali — os dois ajustes juntos é que
 *    tiram de vez a faixa cinza ao lado da câmera.
 *
 * Não verificado por build real neste ambiente (sem Android SDK/Gradle) —
 * ver docs/STATUS.md.
 */
public class MainActivity extends BridgeActivity {

    {
        registerPlugin(SecureTokenStoragePlugin.class);
        registerPlugin(DeviceControlPlugin.class);
    }

    @Override
    public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        getWindow().addFlags(WindowManager.LayoutParams.FLAG_KEEP_SCREEN_ON);
        allowContentUnderCameraCutout();
        WindowCompat.setDecorFitsSystemWindows(getWindow(), false);
    }

    @SuppressWarnings("deprecation")
    private void allowContentUnderCameraCutout() {
        if (Build.VERSION.SDK_INT < Build.VERSION_CODES.P) return;

        WindowManager.LayoutParams params = getWindow().getAttributes();
        params.layoutInDisplayCutoutMode = Build.VERSION.SDK_INT >= Build.VERSION_CODES.R
            ? WindowManager.LayoutParams.LAYOUT_IN_DISPLAY_CUTOUT_MODE_ALWAYS
            : WindowManager.LayoutParams.LAYOUT_IN_DISPLAY_CUTOUT_MODE_SHORT_EDGES;
        getWindow().setAttributes(params);
    }
}
